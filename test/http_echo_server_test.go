package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"strings"
	"test/test/pkg/api"
	"test/test/pkg/test"
	"testing"
)

func TestHttpEchoServer(t *testing.T) {
	var namespaceYaml string = `
apiVersion: v1
kind: Namespace
metadata:
  name: http-echo-server
`
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

				err = api.ApplyAll(*clientSet, objectList)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deployment became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := test.DeploymentBecameReady(ctx, client, "http-echo-server")
				assert.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, install)
}
