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
	err := test.PrepareTest(
		"../kubernetes-services/templates/argo-cd.yaml",
		&argoAppCurrent,
		&argoAppUpdate,
	)

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	install := features.
		New("Deploying Argo CD Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmChart(argoAppCurrent, cfg)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Pods became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.CheckPodsBecameReady(argoAppCurrent)
				assert.NoError(t, err)

				return ctx
			}).
		Assess("Jobs run successfully",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.CheckJobsCompleted(argoAppCurrent, ctx)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()
	upgrade := features.
		New("Upgrading Argo CD Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if argoAppUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			err := test.UpgradeHelmChart(argoAppUpdate, cfg)
			assert.NoError(t, err)

			return ctx
		}).
		Assess("Pods became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.CheckPodsBecameReady(argoAppUpdate)
				assert.NoError(t, err)

				return ctx
			}).
		Assess("Jobs run successfully",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.CheckJobsCompleted(argoAppCurrent, ctx)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
