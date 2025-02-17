package test

import (
	"context"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"strings"
	"test/test/pkg/api"
	"test/test/pkg/test"
	"testing"
)

func TestStorage(t *testing.T) {
	var pvc string = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-csi-pvc
  namespace: default
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Mi
  storageClassName: longhorn
`

	scCurrent, scUpdate, _, err := test.PrepareTest(gitRepository, "../kubernetes-services/templates/snapshot-controller.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare shanpshot controller test #%v", err)
	}

	longhornCurrent, longhornUpdate, manifest, err := test.PrepareTest(gitRepository, "../kubernetes-services/templates/longhorn.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare longhorn csi #%v", err)
	}

	client, err := test.GetClient()
	if err != nil {
		t.Fatalf("Failed to get kubernetes client #%v", err)
	}

	clientSet, err := test.GetClientSet()
	if err != nil {
		t.Fatalf("Failed to get kubernetes clientSet #%v", err)
	}

	feature := features.
		New("Longhorn CSI Test").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {

			// in ci we run a single node instance of kind
			longhornCurrent.Spec.Sources[0].Helm.Values = strings.Replace(
				longhornCurrent.Spec.Sources[0].Helm.Values,
				"defaultClassReplicaCount: 4",
				"defaultClassReplicaCount: 1",
				-1,
			)

			longhornCurrent.Spec.Sources[0].Helm.Values = `
csi:
  attacherReplicaCount: 1
  provisionerReplicaCount: 1
  resizerReplicaCount: 1
  snapshotterReplicaCount: 1
`

			// we also do not have prometheus
			longhornCurrent.Spec.Sources[0].Helm.Values = strings.Replace(
				longhornCurrent.Spec.Sources[0].Helm.Values,
				"serviceMonitor:\n    enabled: true",
				"serviceMonitor:\n    enabled: false",
				-1,
			)

			return ctx
		}).
		Assess("Deploying CSI Helm Charts",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeployHelmCharts(cfg.KubeconfigFile(), scCurrent)
				require.NoError(t, err)

				err = test.DeployHelmCharts(cfg.KubeconfigFile(), longhornCurrent)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Longhorn DaemonSet became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DaemonSetBecameReady(ctx, client, longhornCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, longhornCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Snapshot Controller Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, scCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deploy Snapshot Class",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = api.ApplyAll(*clientSet, manifest)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deploy PVC",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				pvcObject, err := decoder.DecodeAll(ctx, strings.NewReader(pvc))
				err = api.ApplyAll(*clientSet, pvcObject)
				require.NoError(t, err)

				err = test.PersistentVolumeClaimIsBound(ctx, client, "default")
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	upgrade := features.
		New("Upgrading CSI Helm Charts").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if longhornUpdate.Spec.Sources == nil && scUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			if scUpdate.Spec.Sources != nil {
				err = test.DeployHelmCharts(cfg.KubeconfigFile(), scUpdate)
				require.NoError(t, err)
			}

			if longhornUpdate.Spec.Sources != nil {
				// in ci we run a single node instance of kind
				longhornUpdate.Spec.Sources[0].Helm.Values = strings.Replace(
					longhornUpdate.Spec.Sources[0].Helm.Values,
					"defaultClassReplicaCount: 4",
					"defaultClassReplicaCount: 1",
					-1,
				)

				longhornUpdate.Spec.Sources[0].Helm.Values = `
csi:
  attacherReplicaCount: 1
  provisionerReplicaCount: 1
  resizerReplicaCount: 1
  snapshotterReplicaCount: 1
`

				// we also do not have prometheus
				longhornUpdate.Spec.Sources[0].Helm.Values = strings.Replace(
					longhornUpdate.Spec.Sources[0].Helm.Values,
					"serviceMonitor:\n    enabled: true",
					"serviceMonitor:\n    enabled: false",
					-1,
				)

				err = test.DeployHelmCharts(cfg.KubeconfigFile(), longhornUpdate)
				require.NoError(t, err)
			}
			return ctx
		}).
		Assess("Longhorn DaemonSet became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if longhornUpdate.Spec.Sources == nil {
					t.SkipNow()
				}

				err = test.DaemonSetBecameReady(ctx, client, longhornUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Longhorn Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if longhornUpdate.Spec.Sources == nil {
					t.SkipNow()
				}

				err = test.DeploymentBecameReady(ctx, client, longhornUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Snapshot Controller Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if scUpdate.Spec.Sources == nil {
					t.SkipNow()
				}

				err = test.DeploymentBecameReady(ctx, client, scUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, feature, upgrade)
}
