package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"test/test/pkg/helper"
	"test/test/pkg/test"
	"testing"
)

func TestCloudflareTunnel(t *testing.T) {
	var additionalManifests []runtime.Object
	var err error
	err = test.PrepareTest(
		"../kubernetes-services/templates/cloudflare-tunnel.yaml",
		&argoAppCurrent,
		&argoAppUpdate,
		&additionalManifests,
	)

	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	cloudflareTunnelDeployment := additionalManifests[0].(*v1.Deployment)
	cloudflareTunnelDeployment.ObjectMeta.Namespace = argoAppCurrent.Spec.Destination.Namespace

	// delete initContainers, we do not have the tunnel token in ci
	cloudflareTunnelDeployment.Spec.Template.Spec.InitContainers = nil

	// overwrite args with cloudflare hello world test
	cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args = nil
	cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args = append(cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args, "tunnel", "--no-autoupdate", "--hello-world")

	k8sResources, err := mockResources(argoAppCurrent.Spec.Destination.Namespace)
	if err != nil {
		t.Fatalf("Failed to create mock resources #%v", err)
	}

	k8sResources = append(k8sResources, additionalManifests[0])

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to create client #%v", err)
	}

	install := features.
		New("Deploying Cloudflare Tunnel").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for _, object := range k8sResources {
				err = helper.Apply(*client, object)
				if err != nil {
					t.Fatalf("Failed to create object #%v", object)
				}
			}
			require.NoError(t, err)

			return ctx
		}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.DeploymentBecameReady(argoAppCurrent)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install)
}

func mockResources(namespace string) ([]runtime.Object, error) {
	mockNamespace := `
apiVersion: v1
kind: Namespace
metadata:
  name: ` + namespace

	namespaceObject, err := helper.Decode([]byte(mockNamespace))
	if err != nil {
		return nil, err
	}

	resources := make([]runtime.Object, 1)
	resources[0] = namespaceObject

	return resources, nil
}
