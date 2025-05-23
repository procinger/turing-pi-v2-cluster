package test

import (
	"context"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"test/test/pkg/test"
	"testing"
)

func TestCertManager(t *testing.T) {
	current, update, _, err := test.PrepareTest(gitRepository, "../kubernetes-services/templates/cert-manager.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to get kubernetes client #%v", err)
	}

	install := features.
		New("Deploying Cert Manager Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmCharts(cfg.KubeconfigFile(), current)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, current.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	upgrade := features.
		New("Upgrading Cert Manager Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if update.Spec.Sources == nil {
				t.SkipNow()
			}

			err = test.DeployHelmCharts(cfg.KubeconfigFile(), update)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Testing Cert Manager upgrade became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, update.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
