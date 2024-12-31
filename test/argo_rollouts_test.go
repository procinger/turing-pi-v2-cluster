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

func TestArgoRollouts(t *testing.T) {
	err := test.PrepareTest(
		"../kubernetes-services/templates/argo-rollouts.yaml",
		&argoAppCurrent,
		&argoAppUpdate,
	)

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to get kubernetes client #%v", err)
	}

	install := features.
		New("Deploying Argo Rollouts Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmCharts(argoAppCurrent, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.DeploymentBecameReady(ctx, client, argoAppCurrent.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	upgrade := features.
		New("Upgrading Argo Rollouts Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if argoAppUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			err := test.DeployHelmCharts(argoAppUpdate, cfg)
			assert.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.DeploymentBecameReady(ctx, client, argoAppCurrent.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
