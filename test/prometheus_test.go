package e2eutils

import (
	"context"
	"e2eutils/pkg"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestPrometheus(t *testing.T) {
	prometheusTest, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/prometheus.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare prometheus test #%v", err)
	}

	for i, source := range prometheusTest.Current.Argo.Spec.Sources {
		if source.Chart == "" {
			continue
		}
		// replace storageClass; we do not have longhorn in ci
		source.Helm.Values = strings.Replace(
			source.Helm.Values,
			"longhorn",
			"standard",
			-1,
		)

		// also shrink volume
		source.Helm.Values = strings.Replace(
			source.Helm.Values,
			"storage: 20Gi",
			"storage: 500Mi",
			-1,
		)

		source.Helm.Values = strings.Replace(
			source.Helm.Values,
			"storage: 5Gi",
			"storage: 500Mi",
			-1,
		)

		source.Helm.Values = strings.Replace(
			source.Helm.Values,
			"size: 1Gi",
			"size: 100Mi",
			-1,
		)

		prometheusTest.Current.Argo.Spec.Sources[i].Helm.Values = source.Helm.Values
	}

	for i, source := range prometheusTest.Update.Argo.Spec.Sources {
		if source.Chart == "" {
			continue
		}
		// replace storageClass; we do not have longhorn in ci
		source.Helm.Values = strings.Replace(
			source.Helm.Values,
			"longhorn",
			"standard",
			-1,
		)

		// also shrink volume
		source.Helm.Values = strings.Replace(
			source.Helm.Values,
			"storage: 20Gi",
			"storage: 500Mi",
			-1,
		)

		source.Helm.Values = strings.Replace(
			source.Helm.Values,
			"storage: 5Gi",
			"storage: 500Mi",
			-1,
		)

		source.Helm.Values = strings.Replace(
			source.Helm.Values,
			"size: 1Gi",
			"size: 100Mi",
			-1,
		)

		prometheusTest.Update.Argo.Spec.Sources[i].Helm.Values = source.Helm.Values
	}

	client := e2eutils.GetClient()

	install := features.
		New("Deploying Prometheus Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), prometheusTest.Current.Argo)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, prometheusTest.Current.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Daemonsets became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DaemonSetBecameReady(ctx, client, prometheusTest.Current.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	upgrade := features.
		New("Upgrading Prometheus Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if prometheusTest.Update.Argo.Spec.Sources == nil {
				t.SkipNow()
			}

			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), prometheusTest.Update.Argo)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, prometheusTest.Update.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Daemonsets became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DaemonSetBecameReady(ctx, client, prometheusTest.Update.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
