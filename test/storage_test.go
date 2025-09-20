package e2eutils

import (
	"context"
	"e2eutils/pkg"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
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

	snapshotTest, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/snapshot-controller.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare shanpshot controller test: %v", err)
	}

	longhornTest, err := e2eutils.PrepareArgoApp(t.Context(), gitRepository, "../kubernetes-services/templates/longhorn.yaml")
	if err != nil {
		t.Fatalf("Failed to prepare longhorn csi: %v", err)
	}

	client := e2eutils.GetClient()

	clientSet, err := e2eutils.GetClientSet()
	if err != nil {
		t.Fatalf("Failed to get kubernetes clientSet #%v", err)
	}

	feature := features.
		New("Longhorn CSI Test").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {

			// in ci we run a single node instance of kind
			longhornTest.Current.Argo.Spec.Sources[0].Helm.Values = strings.Replace(
				longhornTest.Current.Argo.Spec.Sources[0].Helm.Values,
				"defaultClassReplicaCount: 4",
				"defaultClassReplicaCount: 1",
				-1,
			)

			longhornTest.Current.Argo.Spec.Sources[0].Helm.Values = `
csi:
  attacherReplicaCount: 1
  provisionerReplicaCount: 1
  resizerReplicaCount: 1
  snapshotterReplicaCount: 1
`

			// we also do not have prometheus
			longhornTest.Current.Argo.Spec.Sources[0].Helm.Values = strings.Replace(
				longhornTest.Current.Argo.Spec.Sources[0].Helm.Values,
				"serviceMonitor:\n    enabled: true",
				"serviceMonitor:\n    enabled: false",
				-1,
			)

			return ctx
		}).
		Assess("Deploying CSI Helm Charts",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), snapshotTest.Current.Argo)
				require.NoError(t, err)

				err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), longhornTest.Current.Argo)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Longhorn DaemonSet became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DaemonSetBecameReady(ctx, client, longhornTest.Current.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, longhornTest.Current.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Snapshot Controller Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.DeploymentBecameReady(ctx, client, snapshotTest.Current.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deploy Snapshot Class",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = e2eutils.ApplyAll(*clientSet, longhornTest.Current.Objects)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deploy PVC",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				pvcObject, err := decoder.DecodeAll(ctx, strings.NewReader(pvc))
				err = e2eutils.ApplyAll(*clientSet, pvcObject)
				require.NoError(t, err)

				err = e2eutils.PersistentVolumeClaimIsBound(ctx, client, "default")
				require.NoError(t, err)

				return ctx
			}).
		Feature()
	upgrade := features.
		New("Upgrading CSI Helm Charts").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if longhornTest.Update.Argo.Spec.Sources == nil && snapshotTest.Update.Argo.Spec.Sources == nil {
				t.SkipNow()
			}

			if snapshotTest.Update.Argo.Spec.Sources != nil {
				err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), snapshotTest.Update.Argo)
				require.NoError(t, err)
			}

			if longhornTest.Update.Argo.Spec.Sources != nil {
				// in ci we run a single node instance of kind
				longhornTest.Update.Argo.Spec.Sources[0].Helm.Values = strings.Replace(
					longhornTest.Update.Argo.Spec.Sources[0].Helm.Values,
					"defaultClassReplicaCount: 4",
					"defaultClassReplicaCount: 1",
					-1,
				)

				longhornTest.Update.Argo.Spec.Sources[0].Helm.Values = `
csi:
  attacherReplicaCount: 1
  provisionerReplicaCount: 1
  resizerReplicaCount: 1
  snapshotterReplicaCount: 1
`

				// we also do not have prometheus
				longhornTest.Update.Argo.Spec.Sources[0].Helm.Values = strings.Replace(
					longhornTest.Update.Argo.Spec.Sources[0].Helm.Values,
					"serviceMonitor:\n    enabled: true",
					"serviceMonitor:\n    enabled: false",
					-1,
				)

				err = e2eutils.DeployHelmCharts(cfg.KubeconfigFile(), longhornTest.Update.Argo)
				require.NoError(t, err)
			}
			return ctx
		}).
		Assess("Longhorn DaemonSet became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if longhornTest.Update.Argo.Spec.Sources == nil {
					t.SkipNow()
				}

				err = e2eutils.DaemonSetBecameReady(ctx, client, longhornTest.Update.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Longhorn Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if longhornTest.Update.Argo.Spec.Sources == nil {
					t.SkipNow()
				}

				err = e2eutils.DeploymentBecameReady(ctx, client, longhornTest.Update.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Snapshot Controller Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if snapshotTest.Update.Argo.Spec.Sources == nil {
					t.SkipNow()
				}

				err = e2eutils.DeploymentBecameReady(ctx, client, snapshotTest.Update.Argo.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, feature, upgrade)
}
