package test

import (
	"context"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/yaml"
	"test/test/pkg/api"
	"test/test/pkg/manifest"
	"test/test/pkg/test"
	"testing"
	"time"
)

func TestK3sSystemUpgradeController(t *testing.T) {
	_, update, _, err := test.PrepareTest("../kubernetes-services/templates/k3s-system-upgrade-controller.yaml")
	require.NoError(t, err)

	clientSet, err := test.GetClientSet()
	require.NoError(t, err)

	client, err := test.GetClient()
	require.NoError(t, err)

	var kustomization []string

	install := features.
		New("Kustomization").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			kustomization, err = manifest.BuildKustomization(update.Spec.Sources[0].Path)
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployment",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				for _, resource := range kustomization {
					var object unstructured.Unstructured
					err = yaml.Unmarshal([]byte(resource), &object.Object)
					require.NoError(t, err)

					if object.GetKind() != "Namespace" {
						continue
					}

					err = api.Apply(*clientSet, &object)
					require.NoError(t, err)

					// give k8s api some time to create a resource
					time.Sleep(100 * time.Millisecond)
				}

				for _, resource := range kustomization {
					var object unstructured.Unstructured
					err = yaml.Unmarshal([]byte(resource), &object.Object)
					require.NoError(t, err)

					if object.GetKind() == "Namespace" {
						continue
					}

					err = api.Apply(*clientSet, &object)
					require.NoError(t, err)

					// give k8s api some time to create a resource
					time.Sleep(100 * time.Millisecond)
				}

				return ctx
			}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, "system-upgrade")
				require.NoError(t, err)

				return ctx
			}).
		//		Assess("Plan was applied",
		//			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		//				// Give K3S Plan some time
		//				time.Sleep(60 * time.Second)
		//
		//				pods, err := clientSet.CoreV1().Pods("system-upgrade").List(ctx, metav1.ListOptions{
		//					LabelSelector: "upgrade.cattle.io/plan=k3s-upgrade",
		//				})
		//				require.NoError(t, err)
		//
		//				log, err := clientSet.CoreV1().Pods("system-upgrade").GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{}).Do(ctx).Raw()
		//				require.NoError(t, err)
		//
		//				// If the following error message is thrown, the upgrade plan has been successfully deployed.
		//				// However, it fails because we have Kind instead of K3S here
		//				res := strings.Contains(string(log), "No K3s pids found; is K3s running on this host?")
		//				require.True(t, res)
		//
		//				return ctx
		//			}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err := clientSet.CoreV1().Namespaces().Delete(ctx, "system-upgrade", metav1.DeleteOptions{})
			require.NoError(t, err)
			return ctx
		}).
		Feature()

	ciTestEnv.Test(t, install)
}
