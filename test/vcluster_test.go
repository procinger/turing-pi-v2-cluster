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

func TestVcluster(t *testing.T) {
	current, update, _, err := test.PrepareTest(gitRepository, "../kubernetes-services/templates/vcluster.yaml")
	require.NoError(t, err)

	client, err := test.GetClient()
	require.NoError(t, err)

	install := features.
		New("Deploying Vcluster Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err = test.DeployHelmCharts(cfg.KubeconfigFile(), current)
			require.NoError(t, err)

			return ctx
		}).
		Feature()

	upgrade := features.
		New("Upgrading Vcluster Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if update.Spec.Source == nil {
				t.SkipNow()
			}

			err := test.DeployHelmCharts(cfg.KubeconfigFile(), update)
			assert.NoError(t, err)

			return ctx
		}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
