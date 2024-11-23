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
	err := test.PrepareTest(
		"../kubernetes-services/templates/prometheus.yaml",
		&argoAppCurrent,
		&argoAppUpdate,
	)

	for i, source := range argoAppCurrent.Spec.Sources {
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

		// also shrink volu
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

		if argoAppUpdate.Spec.Sources != nil {
			argoAppUpdate.Spec.Sources[i].Helm.Values = source.Helm.Values
		}
	}

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	feature := features.
		New("Deploying Prometheus Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmCharts(argoAppCurrent, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(argoAppCurrent)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	upgrade := features.
		New("Upgrading Prometheus Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if argoAppUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			err = test.UpgradeHelmChart(argoAppUpdate, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(argoAppUpdate)
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	ciTestEnv.Test(t, feature, upgrade)

}
