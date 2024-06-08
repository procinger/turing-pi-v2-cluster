#!/usr/bin/env bash

_kind_setup() {
   cat <<EOF | kind create cluster --config=- >&2
   kind: Cluster
   apiVersion: kind.x-k8s.io/v1alpha4
   name: ci-test
   featureGates:
     KubeletInUserNamespace: true
   networking:
     ipFamily: ipv4
     apiServerPort: 6443
     apiServerAddress: 127.0.0.1
     podSubnet: "11.244.0.0/16"
   nodes:
     - role: control-plane
       extraMounts:
         # quick hack to fix missing directory in kind
         - hostPath: /var/lib/containerd/io.containerd.snapshotter.v1.fuse-overlayfs
           containerPath: /var/lib/containerd/io.containerd.snapshotter.v1.fuse-overlayfs
EOF
}

_kind_teardown() {
    kind delete cluster --name=ci-test
}

_kind_became_ready() {
  kubectl wait node --all --for condition=ready --timeout=60s
}

_export_helm_values() {
  export CI_COMMIT_SHA=e5ae878b4d9116ba59cb88a6da05fdb0f857afcd
  export RENOVATE_UPDATE=$(git diff-tree --no-commit-id --name-only -r ${CI_COMMIT_SHA})

  export HELM_CHART=$(yq '.spec.source.chart' "${RENOVATE_UPDATE}")
  export HELM_CHART_VERSION=$(yq '.spec.source.targetRevision' "${RENOVATE_UPDATE}")
  export HELM_REPOSITORY=$(yq '.spec.source.repoURL' "${RENOVATE_UPDATE}")
  export NAMESPACE=$(yq '.spec.destination.namespace' "${RENOVATE_UPDATE}")
  yq '.spec.source.helm.values' "${RENOVATE_UPDATE}" > /tmp/values.yaml

  git checkout origin/main ${RENOVATE_UPDATE}
  export MAIN_HELM_CHART_VERSION=$(yq '.spec.source.targetRevision' "${RENOVATE_UPDATE}")
}
