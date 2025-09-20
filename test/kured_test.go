package e2eutils

import (
	"context"
	"e2eutils/pkg"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestKured(t *testing.T) {
	current, update, _, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/kured.yaml")
	require.NoError(t, err)

	client := e2eutils.GetClient()

	install := features.
		New("Deploying Kured Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), current)
			require.NoError(t, err)

			return ctx
		}).
		Assess("DaemonSet became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.DaemonSetBecameReady(ctx, client, current.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	upgrade := features.
		New("Upgrading Kured Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if update.Spec.Source == nil {
				t.SkipNow()
			}

			err := e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), update)
			assert.NoError(t, err)

			return ctx
		}).
		Assess("DaemonSet became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.DaemonSetBecameReady(ctx, client, update.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
