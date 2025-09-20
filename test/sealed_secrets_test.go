package e2eutils

import (
	"context"
	"e2eutils/pkg"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestSealedSecrets(t *testing.T) {
	sealedCurrent, sealedUpdate, _, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/sealed-secrets.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare sealed secret test #%v", err)
	}

	client := e2eutils.GetClient()

	install := features.
		New("Deploying Sealed Secrets Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), sealedCurrent)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, sealedCurrent.Spec.Destination.Namespace)
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

			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), sealedUpdate)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Testing Sealed Secrets upgrade became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, sealedUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
