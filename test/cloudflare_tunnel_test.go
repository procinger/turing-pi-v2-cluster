package e2eutils

import (
	"context"
	"e2eutils/pkg"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestCloudflareTunnel(t *testing.T) {
	var initYaml string = `
apiVersion: v1
kind: Namespace
metadata:
  name: %s

`
	cloudflareTest, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/cloudflare-tunnel.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare test: %v", err)
	}

	client := e2eutils.GetClient()

	clientSet, err := e2eutils.GetClientSet()
	if err != nil {
		t.Fatalf("Failed to get kubernetes clientSet: %v", err)
	}

	var objectList []k8s.Object
	install := features.
		New("Preparing Cloudflare Tunnel Test").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			cloudflareTunnelDeployment := cloudflareTest.Current.Objects[0].(*v1.Deployment)
			cloudflareTunnelDeployment.ObjectMeta.Namespace = cloudflareTest.Current.Argo.Spec.Destination.Namespace

			// delete initContainers, we do not have the tunnel token in ci
			cloudflareTunnelDeployment.Spec.Template.Spec.InitContainers = nil

			// overwrite args with cloudflare hello world test
			cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args = nil
			cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args = append(cloudflareTunnelDeployment.Spec.Template.Spec.Containers[0].Args, "tunnel", "--no-autoupdate", "--hello-world")

			initYaml = fmt.Sprintf(initYaml, "cloudflare")

			objectList, err = decoder.DecodeAll(ctx, strings.NewReader(initYaml))
			if err != nil {
				t.Fatalf("Failed to decode namespace: %v", err)
			}

			objectList = append(objectList, cloudflareTest.Current.Objects[0])

			return ctx
		}).
		Assess("Deploying Cloudflare Tunnel",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if len(objectList) < 1 {
					t.Fatalf("No objects to deploy %v", objectList)
				}

				err = e2eutils.ApplyAll(*clientSet, objectList)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := e2eutils.DeploymentBecameReady(ctx, client, cloudflareTest.Current.Argo.Spec.Destination.Namespace)
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install)
}
