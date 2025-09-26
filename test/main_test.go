package e2eutils

import (
	e2eutils "e2eutils/pkg"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	ciTestEnv     env.Environment
	clusterName   string
	gitRepository string
)

func TestMain(m *testing.M) {
	config, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Printf("Could not create config from env: %v\n", err)
	}

	gitRepository = "https://raw.githubusercontent.com/procinger/turing-pi-v2-cluster/refs/heads/main/"
	gitRepo := "git@github.com:procinger/turing-pi-v2-cluster.git"
	path := filepath.Join(os.TempDir(), "turing-pi-v2-cluster")
	e2eutils.Clone(path, gitRepo)

	ciTestEnv = env.NewWithConfig(config)
	clusterName = envconf.RandomName("ci-e2e-test", 16)

	ciTestEnv.Setup(
		envfuncs.CreateClusterWithConfig(kind.NewProvider(), clusterName, "kind.yaml"),
	)

	ciTestEnv.Finish(
		envfuncs.ExportClusterLogs(clusterName, "./kind-logs"),
		envfuncs.DestroyCluster(clusterName),
	)

	os.Exit(ciTestEnv.Run(m))
}
