package test

import (
	"context"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"test/test/pkg/test"
	"testing"
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
			err = test.DeployHelmCharts(argoAppCurrent, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(argoAppCurrent)
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	upgrade := features.
		New("Upgrading Sealed Secrets Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if argoAppUpdate.Spec.Source == nil {
				t.SkipNow()
			}

			err = test.UpgradeHelmChart(argoAppCurrent, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Testing Sealed Secrets upgrade became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(argoAppUpdate)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
