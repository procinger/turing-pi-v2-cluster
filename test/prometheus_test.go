package test

import (
	"context"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"strings"
	"test/test/pkg/test"
	"testing"
)

func TestPrometheus(t *testing.T) {
	promCurrent, promUpdate, _, err := test.PrepareTest("../kubernetes-services/templates/prometheus.yaml")
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

		if promCurrent.Spec.Sources != nil {
			promCurrent.Spec.Sources[i].Helm.Values = source.Helm.Values
		}
	}

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to get kubernetes client #%v", err)
	}

	install := features.
		New("Deploying Prometheus Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmCharts(cfg.KubeconfigFile(), promCurrent)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, promCurrent.Spec.Destination.Namespace)
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

			err = test.DeployHelmCharts(cfg.KubeconfigFile(), promUpdate)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, promUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	ciTestEnv.Test(t, install, upgrade)

}
