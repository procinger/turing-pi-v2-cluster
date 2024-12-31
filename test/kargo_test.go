package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"test/test/pkg/argo"
	"test/test/pkg/test"
	"testing"
)

var (
	kargoAppCurrent       argo.Application
	kargoAppUpdate        argo.Application
	certManagerAppCurrent argo.Application
	certManagerAppUpdate  argo.Application
)

func TestKargo(t *testing.T) {
	err := test.PrepareTest(
		"../kubernetes-services/templates/kargo.yaml",
		&kargoAppCurrent,
		&kargoAppUpdate,
	)

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	err = test.PrepareTest(
		"../kubernetes-services/templates/cert-manager.yaml",
		&certManagerAppCurrent,
		&certManagerAppUpdate,
	)

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to get kubernetes client #%v", err)
	}

	install := features.
		New("Deploying Kargo.io Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for index, _ := range kargoAppCurrent.Spec.Sources {
				kargoAppCurrent.Spec.Sources[index].RepoURL = "oci://" + kargoAppCurrent.Spec.Sources[index].RepoURL + "/" + kargoAppCurrent.Spec.Sources[index].Chart
			}

			// kargo depends on cert-manager crds
			err = test.DeployHelmCharts(certManagerAppCurrent, cfg)
			require.NoError(t, err)

			err = test.DeployHelmCharts(kargoAppCurrent, cfg)
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
		New("Upgrading Kargo.io Helm Chart").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if kargoAppUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			for index, _ := range kargoAppUpdate.Spec.Sources {
				kargoAppUpdate.Spec.Sources[index].RepoURL = "oci://" + kargoAppUpdate.Spec.Sources[index].RepoURL + "/" + kargoAppUpdate.Spec.Sources[index].Chart
			}

			err := test.DeployHelmCharts(kargoAppUpdate, cfg)
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
