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
	sealedCurrent, sealedUpdate, _, err := test.PrepareTest("../kubernetes-services/templates/sealed-secrets.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare sealed secret test #%v", err)
	}

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to get kubernetes client #%v", err)
	}

	install := features.
		New("Deploying Sealed Secrets Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmCharts(cfg.KubeconfigFile(), sealedCurrent)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, sealedCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	upgrade := features.
		New("Upgrading Sealed Secrets Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if sealedUpdate.Spec.Source == nil {
				t.SkipNow()
			}

			err = test.DeployHelmCharts(cfg.KubeconfigFile(), sealedUpdate)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Testing Sealed Secrets upgrade became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, sealedUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
