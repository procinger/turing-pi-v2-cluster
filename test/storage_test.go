package test

import (
	"context"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"strings"
	"test/test/pkg/api"
	"test/test/pkg/argo"
	"test/test/pkg/test"
	"testing"
)

var (
	snapshotControllerAppCurrent argo.Application
	snapshotControllerAppUpdate  argo.Application
	longhornAppCurrent           argo.Application
	longhornAppUpdate            argo.Application
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

	err := test.PrepareTest(
		"../kubernetes-services/templates/snapshot-controller.yaml",
		&snapshotControllerAppCurrent,
		&snapshotControllerAppUpdate,
	)

	if err != nil {
		t.Fatalf("Failed to prepare shanpshot controller #%v", err)
	}

	var additionalManifests []k8s.Object
	err = test.PrepareTest(
		"../kubernetes-services/templates/longhorn.yaml",
		&longhornAppCurrent,
		&longhornAppUpdate,
		&additionalManifests,
	)

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
			if longhornAppCurrent.Spec.Source != nil {
				// in ci we run a single node instance of kind
				longhornAppCurrent.Spec.Source.Helm.Values = strings.Replace(
					longhornAppCurrent.Spec.Source.Helm.Values,
					"defaultClassReplicaCount: 4",
					"defaultClassReplicaCount: 1",
					-1,
				)

				longhornAppCurrent.Spec.Source.Helm.Values = `
csi:
  attacherReplicaCount: 1
  provisionerReplicaCount: 1
  resizerReplicaCount: 1
  snapshotterReplicaCount: 1
`

				// we also do not have prometheus
				longhornAppCurrent.Spec.Source.Helm.Values = strings.Replace(
					longhornAppCurrent.Spec.Source.Helm.Values,
					"serviceMonitor:\n    enabled: true",
					"serviceMonitor:\n    enabled: false",
					-1,
				)
			}

			return ctx
		}).
		Assess("Deploying CSI Helm Charts",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeployHelmCharts(snapshotControllerAppCurrent, cfg)
				require.NoError(t, err)

				err = test.DeployHelmCharts(longhornAppCurrent, cfg)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Longhorn DaemonSet became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DaemonSetBecameReady(ctx, client, argoAppCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = test.DeploymentBecameReady(ctx, client, argoAppCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				err = test.DeploymentBecameReady(ctx, client, argoAppCurrent.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Deploy Snapshot Class",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err = api.ApplyAll(*clientSet, additionalManifests)
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
			if longhornAppUpdate.Spec.Sources == nil && snapshotControllerAppUpdate.Spec.Sources == nil {
				t.SkipNow()
			}

			// in ci we run a single node instance of kind
			longhornAppUpdate.Spec.Sources[0].Helm.Values = strings.Replace(
				longhornAppUpdate.Spec.Sources[0].Helm.Values,
				"defaultClassReplicaCount: 4",
				"defaultClassReplicaCount: 1",
				-1,
			)

			longhornAppUpdate.Spec.Sources[0].Helm.Values = `
csi:
  attacherReplicaCount: 1
  provisionerReplicaCount: 1
  resizerReplicaCount: 1
  snapshotterReplicaCount: 1
`

			// we also do not have prometheus
			longhornAppUpdate.Spec.Sources[0].Helm.Values = strings.Replace(
				longhornAppUpdate.Spec.Sources[0].Helm.Values,
				"serviceMonitor:\n    enabled: true",
				"serviceMonitor:\n    enabled: false",
				-1,
			)

			if snapshotControllerAppUpdate.Spec.Sources != nil {
				err = test.DeployHelmCharts(snapshotControllerAppUpdate, cfg)
				require.NoError(t, err)
			}

			if longhornAppUpdate.Spec.Sources != nil {
				err = test.DeployHelmCharts(longhornAppUpdate, cfg)
				require.NoError(t, err)
			}
			return ctx
		}).
		Assess("Longhorn DaemonSet became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if longhornAppUpdate.Spec.Sources == nil {
					t.SkipNow()
				}

				err = test.DaemonSetBecameReady(ctx, client, longhornAppUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Longhorn Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if longhornAppUpdate.Spec.Sources == nil {
					t.SkipNow()
				}

				err = test.DeploymentBecameReady(ctx, client, longhornAppUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Assess("Snapshot Controller Deployments became ready",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if snapshotControllerAppUpdate.Spec.Sources == nil {
					t.SkipNow()
				}

				err = test.DeploymentBecameReady(ctx, client, snapshotControllerAppUpdate.Spec.Destination.Namespace)
				require.NoError(t, err)

				return ctx
			}).
		Feature()

	ciTestEnv.Test(t, feature, upgrade)
}
