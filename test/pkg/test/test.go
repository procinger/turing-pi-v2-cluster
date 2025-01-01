package test

import (
	"context"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log/slog"
	"reflect"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"strings"
	"test/test/pkg/argo"
	"test/test/pkg/git"
	"test/test/pkg/helm"
	"test/test/pkg/manifest"
	"time"
)

func PrepareTest(
	applicationYaml string,
) (argo.Application, argo.Application, []k8s.Object, error) {
	currGitBranch, err := git.GetCurrentGitBranch()
	if err != nil {
		return argo.Application{}, argo.Application{}, nil, err
	}

	if currGitBranch == "main" {
		current, err := argo.GetArgoApplication(applicationYaml)
		if err != nil {
			return argo.Application{}, argo.Application{}, nil, err
		}

		objects, err := manifest.GetKubernetesManifests(current)
		if err != nil {
			return current, argo.Application{}, nil, err
		}

		return current, argo.Application{}, objects, nil
	}

	current, err := argo.GetArgoApplicationFromGit(applicationYaml)
	if err != nil {
		return argo.Application{}, argo.Application{}, nil, err
	}

	update, err := argo.GetArgoApplication(applicationYaml)
	if err != nil {
		return argo.Application{}, argo.Application{}, nil, err
	}

	objects, err := manifest.GetKubernetesManifests(update)
	if err != nil {
		return argo.Application{}, argo.Application{}, nil, err
	}

	if current.Spec.Source == nil && update.Spec.Sources == nil {
		current = update
		update = argo.Application{}
	}

	if reflect.DeepEqual(current, update) {
		update = argo.Application{}
	}

	return current, update, objects, nil
}

func deployHelmChart(applicationSource argo.ApplicationSource, namespace string, kubeConfigFile string) error {
	helmMgr := helm.GetHelmManager(kubeConfigFile)

	if !strings.Contains(applicationSource.RepoURL, "oci://") {
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

func DeployHelmCharts(kubeConfigFile string, argoApplication argo.Application) error {
	if argoApplication.Spec.Source != nil {
		if argoApplication.Spec.Source.Chart == "" {
			return nil
		}

		err := deployHelmChart(*argoApplication.Spec.Source, argoApplication.Spec.Destination.Namespace, kubeConfigFile)
		if err != nil {
			slog.Error(err.Error())
			return err
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
			slog.Error(err.Error())
			return err
		}
	}

	return nil
}

func GetClient() (klient.Client, error) {
	cfg := envconf.Config{}
	return cfg.Client(), nil
}

func GetRestConfig() *rest.Config {
	client, _ := GetClient()
	return client.RESTConfig()
}

func GetClientSet() (*kubernetes.Clientset, error) {
	kubeConfig := GetRestConfig()
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func GetDynClient() (*dynamic.DynamicClient, error) {
	kubeConfig := GetRestConfig()
	return dynamic.NewForConfig(kubeConfig)
}

func GetDiscoveryClient() (*discovery.DiscoveryClient, error) {
	kubeConfig := GetRestConfig()
	return discovery.NewDiscoveryClientForConfig(kubeConfig)
}

func CheckJobsCompleted(ctx context.Context, client klient.Client, namespace string) error {
	kubeConfig := client.RESTConfig()
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
	kubeConfig := client.RESTConfig()
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
	kubeConfig := client.RESTConfig()
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	daemonSetList, err := clientSet.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
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
	kubeConfig := client.RESTConfig()
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
	kubeConfig := client.RESTConfig()
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
		slog.Error("Failed to list VolumeSnapshots: " + err.Error())
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
