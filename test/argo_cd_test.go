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

func TestArgoCd(t *testing.T) {
	current, update, _, err := e2eutils.PrepareArgoApp(gitRepository, "../kubernetes-services/templates/argo-cd.yaml")

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	client := e2eutils.GetClient()

	install := features.
		New("Deploying Argo CD Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), current)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.DeploymentBecameReady(ctx, client, current.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Assess("Jobs run successfully",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.CheckJobsCompleted(ctx, client, current.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	upgrade := features.
		New("Upgrading Argo CD Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if update.Spec.Sources == nil {
				t.SkipNow()
			}

			err := e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), update)
			assert.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.DeploymentBecameReady(ctx, client, update.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Assess("Jobs run successfully",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.CheckJobsCompleted(ctx, client, update.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
