package e2eutils

import (
	"context"
	"e2eutils/pkg/argo"
	"e2eutils/pkg/helm"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

func PrepareArgoApp(gitRepository string, applicationYaml string) (argo.Application, argo.Application, []k8s.Object, error) {
	currGitBranch, err := GetCurrentGitBranch()
	if err != nil {
		return argo.Application{}, argo.Application{}, nil, err
	}

	if currGitBranch == "main" {
		current, err := argo.GetArgoApplication(applicationYaml)
		if err != nil {
			return argo.Application{}, argo.Application{}, nil, err
		}

		objects, err := GetKubernetesManifests(current)
		if err != nil {
			return current, argo.Application{}, nil, err
		}

		return current, argo.Application{}, objects, nil
	}

	update, err := argo.GetArgoApplication(applicationYaml)
	if err != nil {
		return argo.Application{}, argo.Application{}, nil, err
	}

	objects, err := GetKubernetesManifests(update)
	if err != nil {
		return argo.Application{}, argo.Application{}, nil, err
	}

	current, err := argo.GetArgoApplicationFromGit(gitRepository, applicationYaml)
	if err != nil {
		slog.Warn(
			"Failed to get current application from git",
			"application", applicationYaml,
			"branch", currGitBranch,
			"error", err.Error(),
		)
		return update, argo.Application{}, objects, nil
	}

	if reflect.DeepEqual(current, update) {
		return current, argo.Application{}, objects, nil
	}

	return current, update, objects, nil
}

func DeployHelmCharts(kubeConfigFile string, argoApplication argo.Application) error {
	if argoApplication.Spec.Source != nil && argoApplication.Spec.Source.Chart != "" {
		err := deployHelmChart(*argoApplication.Spec.Source, argoApplication.Spec.Destination.Namespace, kubeConfigFile)
		if err != nil {
			return errors.New(err.Error())
		}

		return nil
	}

	var source argo.ApplicationSource
	for _, source = range argoApplication.Spec.Sources {
		if source.Chart == "" {
			continue
		}

		err := deployHelmChart(source, argoApplication.Spec.Destination.Namespace, kubeConfigFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func deployHelmChart(applicationSource argo.ApplicationSource, namespace string, kubeConfigFile string) error {
	helmMgr := helm.NewHelmManager(kubeConfigFile)

	if !strings.HasPrefix(applicationSource.RepoURL, "oci://") {
		err := helm.AddHelmRepository(helmMgr, applicationSource.RepoURL, applicationSource.Chart)
		if err != nil {
			return err
		}
	}

	err := helm.DeployHelmChart(helmMgr, applicationSource, namespace)
	if err != nil {
		return err
	}

	return nil
}

func CheckJobsCompleted(ctx context.Context, client klient.Client, namespace string) error {
	kubeConfig := GetRestConfig()
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	jobsList, err := clientSet.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range jobsList.Items {
		err = wait.For(
			conditions.New(client.Resources().WithNamespace(namespace)).
				JobCompleted(&jobsList.Items[i]), wait.WithTimeout(time.Minute*10),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func DeploymentBecameReady(ctx context.Context, client klient.Client, namespace string) error {
	kubeConfig := GetRestConfig()
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	deploymentList, err := clientSet.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range deploymentList.Items {
		var isDeploymentDone = func(object k8s.Object) bool {
			dep := object.(*appsv1.Deployment)
			return dep.Status.Replicas == dep.Status.ReadyReplicas
		}

		err = wait.For(
			conditions.New(client.Resources()).ResourceMatch(&deploymentList.Items[i], isDeploymentDone),
			wait.WithTimeout(time.Minute*5),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func DaemonSetBecameReady(ctx context.Context, client klient.Client, namespace string) error {
	kubeConfig := GetRestConfig()
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	daemonSetList, err := clientSet.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range daemonSetList.Items {
		var isDaemonSetDone = func(object k8s.Object) bool {
			dep := object.(*appsv1.DaemonSet)
			return dep.Status.DesiredNumberScheduled == dep.Status.NumberReady
		}

		err = wait.For(
			conditions.New(client.Resources()).ResourceMatch(&daemonSetList.Items[i], isDaemonSetDone),
			wait.WithTimeout(time.Minute*5),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func PersistentVolumeClaimIsBound(ctx context.Context, client klient.Client, namespace string) error {
	kubeConfig := GetRestConfig()
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	pvcList, err := clientSet.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range pvcList.Items {
		var isBound = func(object k8s.Object) bool {
			dep := object.(*corev1.PersistentVolumeClaim)
			return dep.Status.Phase == corev1.ClaimBound
		}

		err = wait.For(
			conditions.New(client.Resources()).ResourceMatch(&pvcList.Items[i], isBound),
			wait.WithTimeout(time.Minute*10),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func SnapshotIsReadyToUse(ctx context.Context, client klient.Client, namespace string) error {
	kubeConfig := GetRestConfig()
	dynClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	gvr := schema.GroupVersionResource{
		Group:    "snapshot.storage.k8s.io",
		Version:  "v1",
		Resource: "volumesnapshots",
	}

	snapshotList, err := dynClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range snapshotList.Items {
		var isReadyToUse = func(object k8s.Object) bool {
			unstructuredObj, ok := object.(*unstructured.Unstructured)
			if !ok {
				return false
			}

			volumeSnapshot := &snapshotv1.VolumeSnapshot{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, volumeSnapshot)
			if err != nil {
				return false
			}

			return *volumeSnapshot.Status.ReadyToUse == true
		}

		err = wait.For(
			conditions.New(client.Resources()).ResourceMatch(&snapshotList.Items[i], isReadyToUse),
			wait.WithTimeout(time.Minute*10),
		)

		if err != nil {
			return err
		}
	}

	return nil
}
