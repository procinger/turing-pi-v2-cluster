package test

import (
	"context"
	"fmt"
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"strings"
	"test/test/pkg/helper"
	"time"
)

func PrepareTest(applicationYaml string, argoAppCurrent *applicationV1Alpha1.Application, argoAppUpdate *applicationV1Alpha1.Application) error {
	currGitBranch, err := helper.GetCurrentGitBranch()
	if err != nil {
		return err
	}

	if currGitBranch == "main" {
		*argoAppCurrent, err = helper.GetArgoApplication(applicationYaml)
		if err != nil {
			return err
		}

		return nil
	}

	*argoAppCurrent, err = helper.GetArgoApplicationFromGit(applicationYaml)
	if err != nil {
		return err
	}

	*argoAppUpdate, err = helper.GetArgoApplication(applicationYaml)
	if err != nil {
		return err
	}

	if argoAppCurrent.Spec.Source == nil && argoAppCurrent.Spec.Sources == nil {
		*argoAppCurrent = *argoAppUpdate
		*argoAppUpdate = applicationV1Alpha1.Application{}
		return nil
	}

	if reflect.DeepEqual(argoAppCurrent, argoAppUpdate) {
		*argoAppUpdate = applicationV1Alpha1.Application{}
	}

	return nil
}

func deployHelmChart(applicationSource applicationV1Alpha1.ApplicationSource,namespace string, cfg *envconf.Config) error {
	helmMgr := helper.GetHelmManager(cfg)

	if ! strings.Contains(applicationSource.RepoURL, "oci://") {
		err := helper.AddHelmRepository(helmMgr, applicationSource.RepoURL, applicationSource.Chart)
		if err != nil {
			return err
		}
	}

	err := helper.InstallHelmChart(helmMgr, applicationSource, namespace)
	if err != nil {
		return err
	}


	return nil
}

func DeployHelmCharts(argoApplication applicationV1Alpha1.Application,  cfg *envconf.Config) error {
	if argoApplication.Spec.Source != nil {
		if argoApplication.Spec.Source.Chart == "" {
			return nil
		}

		err := deployHelmChart(*argoApplication.Spec.Source, argoApplication.Spec.Destination.Namespace, cfg)
		if err != nil {
			return err
		}

		return nil
	}

	var source applicationV1Alpha1.ApplicationSource
	for _, source = range argoApplication.Spec.Sources {
		if source.Chart == "" {
			continue
		}

		err := deployHelmChart(source, argoApplication.Spec.Destination.Namespace, cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpgradeHelmChart(argoApplication applicationV1Alpha1.Application, cfg *envconf.Config) error {
	helmMgr := helper.GetHelmManager(cfg)

	var source applicationV1Alpha1.ApplicationSource
	for _, source = range argoApplication.Spec.Sources {
		if source.Chart == "" {
			continue
		}

		err := helper.UpgradeHelmChart(helmMgr, source, argoApplication.Spec.Destination.Namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func getClient() (*kubernetes.Clientset, error) {
	cfg := envconf.Config{}
	kubeConfig := cfg.Client().RESTConfig()
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func CheckPodsBecameReady(argoApplication applicationV1Alpha1.Application) error {
	cfg := envconf.Config{}
	podList := corev1.PodList{}
	var source applicationV1Alpha1.ApplicationSource

	for _, source = range argoApplication.Spec.Sources {
		if source.Chart == "" {
			continue
		}

		err := cfg.Client().Resources(argoApplication.Spec.Destination.Namespace).
			List(context.TODO(), &podList, resources.WithLabelSelector(
				labels.FormatLabels(map[string]string{
					"helm.sh/chart": fmt.Sprintf("%s-%s",
						source.Chart,
						source.TargetRevision,
					),
				})),
			)

		if err != nil {
			return err
		}

		for i := range podList.Items {
			if podList.Items[i].OwnerReferences[0].Kind == "Job" {
				continue
			}

			err = wait.For(
				conditions.New(cfg.Client().Resources().WithNamespace(argoApplication.Spec.Destination.Namespace)).
					PodReady(&podList.Items[i]), wait.WithTimeout(time.Minute*10),
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CheckJobsCompleted(argoApplication applicationV1Alpha1.Application, ctx context.Context) error {
	clientSet, err := getClient()
	if err != nil {
		return err
	}

	jobsList, err := clientSet.BatchV1().Jobs(argoApplication.Spec.Destination.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	cfg := envconf.Config{}
	for i := range jobsList.Items {
		err = wait.For(
			conditions.New(cfg.Client().Resources().WithNamespace(argoApplication.Spec.Destination.Namespace)).
				JobCompleted(&jobsList.Items[i]), wait.WithTimeout(time.Minute*10),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func DeploymentBecameReady(argoApplication applicationV1Alpha1.Application) error {
	clientSet, err := getClient()
	if err != nil {
		return err
	}

	deploymentList, err := clientSet.AppsV1().Deployments(argoApplication.Spec.Destination.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range deploymentList.Items {
		var isDeploymentDone = func(object k8s.Object) bool {
			dep := object.(*appsv1.Deployment)
			return dep.Status.Replicas == dep.Status.ReadyReplicas
		}

		cfg := envconf.Config{}
		err = wait.For(
			conditions.New(cfg.Client().Resources()).ResourceMatch(&deploymentList.Items[i], isDeploymentDone),
			wait.WithTimeout(time.Minute*5),
		)

		if err != nil {
			return err
		}
	}

	return nil
}