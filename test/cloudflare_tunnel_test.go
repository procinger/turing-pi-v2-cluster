package test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"strings"
	"test/test/pkg/api"
	"test/test/pkg/test"
	"testing"
)

func TestCloudflareTunnel(t *testing.T) {
	var initYaml string = `
apiVersion: v1
kind: Namespace
metadata:
  name: %s

`
	current, _, additionalManifests, err := test.PrepareTest("../kubernetes-services/templates/cloudflare-tunnel.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare test #%v", err)
	}

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to get kubernetes client #%v", err)
	}

	clientSet, err := test.GetClientSet()
	if err != nil {
		t.Fatalf("Failed to get kubernetes clientSet #%v", err)
	}

	var objectList []k8s.Object
	install := features.
		New("Preparing Cloudflare Tunnel Test").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			cloudflareTunnelDeployment := additionalManifests[0].(*v1.Deployment)
			cloudflareTunnelDeployment.ObjectMeta.Namespace = current.Spec.Destination.Namespace

			// delete initContainers, we do not have the tunnel token in ci
			cloudflareTunnelDeployment.Spec.Template.Spec.InitContainers = nil

			// overwrite args with cloudflare hello world test
			cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args = nil
			cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args = append(cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args, "tunnel", "--no-autoupdate", "--hello-world")

			initYaml = fmt.Sprintf(initYaml, "cloudflare")

			objectList, err = decoder.DecodeAll(ctx, strings.NewReader(initYaml))
			if err != nil {
				t.Fatalf("Failed to decode namespace #%v", err)
			}

			objectList = append(objectList, additionalManifests[0])

			return ctx
		}).
		Assess("Deploying Cloudflare Tunnel",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if len(objectList) < 1 {
					t.Fatalf("No objects to deploy %v", objectList)
				}

				err = api.ApplyAll(*clientSet, objectList)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.DeploymentBecameReady(ctx, client, current.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install)
}
