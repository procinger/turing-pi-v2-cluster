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
	promCurrent, promUpdate, _, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/prometheus.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare prometheus test #%v", err)
	}

	for i, source := range promCurrent.Spec.Sources {
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

		promCurrent.Spec.Sources[i].Helm.Values = source.Helm.Values
	}

	for i, source := range promUpdate.Spec.Sources {
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

		promUpdate.Spec.Sources[i].Helm.Values = source.Helm.Values
	}

	client := e2eutils.GetClient()

	install := features.
		New("Deploying Prometheus Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), promCurrent)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, promCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Daemonsets became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DaemonSetBecameReady(ctx, client, promCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	upgrade := features.
		New("Upgrading Prometheus Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if promUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), promUpdate)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, promUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Daemonsets became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DaemonSetBecameReady(ctx, client, promUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
