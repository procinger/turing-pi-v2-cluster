package test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"test/test/pkg/helper"
	"test/test/pkg/test"
	"testing"
	"time"
)

func TestSealedSecrets(t *testing.T) {
	err := test.PrepareTest(
		"../kubernetes-services/templates/sealed-secrets.yaml",
		&argoAppCurrent,
		&argoAppUpdate,
	)
	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	install := features.
		New("Deploying Sealed Secrets Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			helmMgr := helper.GetHelmManager(cfg)

			err := helper.AddHelmRepository(helmMgr, argoAppCurrent.Spec.Source.RepoURL)
			require.NoError(t, err)

			err = helper.InstallHelmChart(helmMgr, *argoAppCurrent.Spec.Source, argoAppCurrent.Spec.Destination.Namespace)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      argoAppCurrent.Spec.Source.Chart,
						Namespace: argoAppCurrent.Spec.Destination.Namespace,
					},
				}

				var isDeploymentDone = func(object k8s.Object) bool {
					dep := object.(*appsv1.Deployment)
					return dep.Status.AvailableReplicas == dep.Status.ReadyReplicas
				}

				err := wait.For(
					conditions.New(cfg.Client().Resources()).ResourceMatch(deployment, isDeploymentDone),
					wait.WithTimeout(time.Minute*5),
				)
				assert.NoError(t, err, "Failed to deploy")

				return ctx
			}).
		Assess("Pod became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				podList := corev1.PodList{}

				err = cfg.Client().Resources(argoAppCurrent.Spec.Destination.Namespace).
					List(context.TODO(), &podList, resources.WithLabelSelector(
					labels.FormatLabels(map[string]string{
						"helm.sh/chart": fmt.Sprintf("%s-%s",
							argoAppCurrent.Spec.Source.Chart,
							argoAppCurrent.Spec.Source.TargetRevision,
						),
					})),
				)

				for i := range podList.Items {
					err = wait.For(
						conditions.New(cfg.Client().Resources().WithNamespace(argoAppCurrent.Spec.Destination.Namespace)).
							PodReady(&podList.Items[i]), wait.WithTimeout(time.Minute*10),
					)
				}

				assert.NoError(t, err, "Failed to start sealed secrets")

				return ctx
			}).
		Feature()
	upgrade := features.
		New("Upgrading Sealed Secrets Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if argoAppUpdate.Spec.Source == nil {
				t.SkipNow()
			}

			if argoAppUpdate.Spec.Source.TargetRevision == argoAppCurrent.Spec.Source.TargetRevision {
				t.SkipNow()
			}

			helmMgr := helper.GetHelmManager(cfg)

			err = helper.UpgradeHelmChart(helmMgr, *argoAppUpdate.Spec.Source, argoAppUpdate.Spec.Destination.Namespace)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Testing Sealed Secrets upgrade became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      argoAppUpdate.Spec.Source.Chart,
						Namespace: argoAppUpdate.Spec.Destination.Namespace,
					},
				}

				var isDeploymentDone = func(object k8s.Object) bool {
					dep := object.(*appsv1.Deployment)
					return dep.Status.AvailableReplicas == dep.Status.ReadyReplicas
				}

				err := wait.For(
					conditions.New(cfg.Client().Resources()).ResourceMatch(deployment, isDeploymentDone),
					wait.WithTimeout(time.Minute*5),
				)
				assert.NoError(t, err, "Failed to deploy")

				return ctx
			}).
		Assess("Pod became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				podList := corev1.PodList{}

				err = cfg.Client().Resources(argoAppCurrent.Spec.Destination.Namespace).
					List(context.TODO(), &podList, resources.WithLabelSelector(
						labels.FormatLabels(map[string]string{
							"helm.sh/chart": fmt.Sprintf("%s-%s",
								argoAppUpdate.Spec.Source.Chart,
								argoAppUpdate.Spec.Source.TargetRevision,
							),
						})),
					)

				for i := range podList.Items {
					err = wait.For(
						conditions.New(cfg.Client().Resources().WithNamespace(argoAppCurrent.Spec.Destination.Namespace)).
							PodReady(&podList.Items[i]), wait.WithTimeout(time.Minute*10),
					)
				}

				assert.NoError(t, err, "Failed to start sealed secrets")

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
