package e2eutils

import (
	"context"
	"e2eutils/pkg"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestK3sSystemUpgradeController(t *testing.T) {
	current, update, _, err := e2eutils.PrepareArgoApp(gitRepository, "../kubernetes-services/templates/k3s-system-upgrade-controller.yaml")
	require.NoError(t, err)

	clientSet, err := e2eutils.GetClientSet()
	require.NoError(t, err)

	client := e2eutils.GetClient()

	var kustomization []*unstructured.Unstructured
	var namespace string

	install := features.
		New("Kustomization").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if update.Spec.Sources != nil {
				kustomization, err = e2eutils.BuildKustomization("../" + update.Spec.Sources[0].Path)
			} else {
				kustomization, err = e2eutils.BuildKustomization("../" + current.Spec.Sources[0].Path)
			}
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployment",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				// deploy namespace
				for _, object := range kustomization {
					if object.GetKind() != "Namespace" {
						continue
					}

					err = e2eutils.Apply(*clientSet, object)
					require.NoError(t, err)

					// give k8s api some time to create a resource
					time.Sleep(100 * time.Millisecond)
				}

				for _, object := range kustomization {
					if object.GetKind() == "Namespace" {
						continue
					}

					err = e2eutils.Apply(*clientSet, object)
					require.NoError(t, err)

					// give k8s api some time to create a resource
					time.Sleep(100 * time.Millisecond)
				}

				return ctx
			}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install)
}
