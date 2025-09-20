package e2eutils

import (
	"context"
	"e2eutils/pkg"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestIstio(t *testing.T) {
	istioTest, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/istio.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare test: %v", err)
	}

	gatewayTest, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/istio-gateway.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	client := e2eutils.GetClient()

	install := features.
		New("Deploying Istio Helm Charts Collection").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), istioTest.Current.Argo)
			require.NoError(t, err)

			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), gatewayTest.Current.Argo)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, istioTest.Current.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				err = e2eutils.DeploymentBecameReady(ctx, client, gatewayTest.Current.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).

		// https://istio.io/latest/docs/setup/upgrade/helm/#canary-upgrade-recommended
		Assess("Migrate Istio CRDs",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				kubeConfig := cfg.Client().RESTConfig()
				dynClient, err := dynamic.NewForConfig(kubeConfig)

				crdGVR := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
				crdsChartIstio, err := dynClient.Resource(crdGVR).List(context.TODO(), metav1.ListOptions{
					LabelSelector: "chart=istio",
				})

				if err != nil {
					log.Fatalf("error fetching crds 'chart=istio': %v", err)
				}

				crdsPartOfIstio, err := dynClient.Resource(crdGVR).List(context.TODO(), metav1.ListOptions{
					LabelSelector: "app.kubernetes.io/part-of=istio",
				})
				if err != nil {
					log.Fatalf("error fetching CRD 'app.kubernetes.io/part-of=istio': %v", err)
				}

				crds := append(crdsChartIstio.Items, crdsPartOfIstio.Items...)
				for _, crd := range crds {
					crdName := crd.GetName()

					crd.SetLabels(map[string]string{"app.kubernetes.io/managed-by": "Helm"})
					crd.SetAnnotations(map[string]string{
						"meta.helm.sh/release-name":      "base",
						"meta.helm.sh/release-namespace": "istio-system",
					})

					_, err := dynClient.Resource(crdGVR).Update(context.TODO(), &crd, metav1.UpdateOptions{})
					if err != nil {
						log.Printf("failed to update CRD %s: %v", crdName, err)
					}
				}
				return ctx
			}).
		Feature()

	upgrade := features.
		New("Upgrading Istio Helm Charts Collection").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if istioTest.Update.Argo.Spec.Sources == nil && gatewayTest.Update.Argo.Spec.Sources == nil {
				t.SkipNow()
			}

			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), istioTest.Update.Argo)
			require.NoError(t, err)

			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), gatewayTest.Update.Argo)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, istioTest.Update.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				err = e2eutils.DeploymentBecameReady(ctx, client, gatewayTest.Update.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	ciTestEnv.Test(t, install, upgrade)

}
