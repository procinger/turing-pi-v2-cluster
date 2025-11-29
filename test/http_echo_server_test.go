package e2eutils

import (
	"context"
	"e2eutils/pkg"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestHttpEchoServer(t *testing.T) {
	var namespaceYaml string = `
apiVersion: v1
kind: Namespace
metadata:
  name: http-echo-server
`
	client := e2eutils.GetClient()
	clientSet, err := e2eutils.GetClientSet()
	require.NoError(t, err)

	var objectList []k8s.Object
	install := features.
		New("Preparing HTTP Echo Server Test").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			deploymentYaml, err := os.ReadFile("../kubernetes-services/additions/http-echo-server/deployment.yaml")
			require.NoError(t, err)

			objectList, err = decoder.DecodeAll(ctx, strings.NewReader(namespaceYaml))
			require.NoError(t, err)

			deployment, err := decoder.DecodeAll(ctx, strings.NewReader(string(deploymentYaml)))
			require.NoError(t, err)

			objectList = append(objectList, deployment[0])

			return ctx
		}).
		Assess("Deploying HTTP Echo Server",
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
				err := e2eutils.DeploymentBecameReady(ctx, client, "http-echo-server")
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install)
}
