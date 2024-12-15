package test

import (
	"context"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"test/test/pkg/test"
	"test/test/pkg/types/argocd"
	"testing"
)

var (
	istioAppCurrent        argocd.Application
	istioAppUpdate         argocd.Application
	istioGatewayAppCurrent argocd.Application
	istioGatewayAppUpdate  argocd.Application
)

func TestIstio(t *testing.T) {
	err := test.PrepareTest(
		"../kubernetes-services/templates/istio.yaml",
		&istioAppCurrent,
		&istioAppUpdate,
	)

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	err = test.PrepareTest(
		"../kubernetes-services/templates/istio-gateway.yaml",
		&istioGatewayAppCurrent,
		&istioGatewayAppUpdate,
	)

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	feature := features.
		New("Deploying Istio Helm Charts Collection").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmCharts(istioAppCurrent, cfg)
			require.NoError(t, err)

			err = test.DeployHelmCharts(istioGatewayAppCurrent, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(istioAppCurrent)
				require.NoError(t, err)

				err = test.DeploymentBecameReady(istioGatewayAppCurrent)
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
			if istioAppUpdate.Spec.Sources == nil && istioGatewayAppUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			err = test.DeployHelmCharts(istioAppUpdate, cfg)
			require.NoError(t, err)

			err = test.DeployHelmCharts(istioGatewayAppUpdate, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(istioAppUpdate)
				require.NoError(t, err)

				err = test.DeploymentBecameReady(istioGatewayAppUpdate)
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	ciTestEnv.Test(t, feature, upgrade)

}
