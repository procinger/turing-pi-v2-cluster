package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"test/test/pkg/test"
	"testing"
)

func TestArgoCd(t *testing.T) {
	current, update, _, err := test.PrepareTest("../kubernetes-services/templates/argo-cd.yaml")

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to get kubernetes client #%v", err)
	}

	install := features.
		New("Deploying Argo CD Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmCharts(current, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.DeploymentBecameReady(ctx, client, current.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Assess("Jobs run successfully",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.CheckJobsCompleted(ctx, client, current.Spec.Destination.Namespace)
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

			err := test.DeployHelmCharts(update, cfg)
			assert.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.DeploymentBecameReady(ctx, client, update.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Assess("Jobs run successfully",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.CheckJobsCompleted(ctx, client, update.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
