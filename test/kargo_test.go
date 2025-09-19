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

func TestKargo(t *testing.T) {
	kargoCurrent, kargoUpdate, _, err := e2eutils.PrepareArgoApp(gitRepository, "../kubernetes-services/templates/kargo.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare kargo test #%v", err)
	}

	certCurrent, _, _, err := e2eutils.PrepareArgoApp(gitRepository, "../kubernetes-services/templates/cert-manager.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare cert-manager #%v", err)
	}

	client := e2eutils.GetClient()

	install := features.
		New("Deploying Kargo.io Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for index, _ := range kargoCurrent.Spec.Sources {
				kargoCurrent.Spec.Sources[index].RepoURL = "oci://" + kargoCurrent.Spec.Sources[index].RepoURL + "/" + kargoCurrent.Spec.Sources[index].Chart
			}

			// kargo depends on cert-manager crds
			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), certCurrent)
			require.NoError(t, err)

			err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), kargoCurrent)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.DeploymentBecameReady(ctx, client, kargoCurrent.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()
	upgrade := features.
		New("Upgrading Kargo.io Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if kargoUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			for index, _ := range kargoUpdate.Spec.Sources {
				kargoUpdate.Spec.Sources[index].RepoURL = "oci://" + kargoUpdate.Spec.Sources[index].RepoURL + "/" + kargoUpdate.Spec.Sources[index].Chart
			}

			err := e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), kargoUpdate)
			assert.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.DeploymentBecameReady(ctx, client, kargoUpdate.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install, upgrade)
}
