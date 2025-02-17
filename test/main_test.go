package test

import (
	"fmt"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
	"testing"
)

var (
	ciTestEnv     env.Environment
	clusterName   string
	gitRepository string
)

func TestMain(m *testing.M) {
	config, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Println("Could not create config from env", err)
	}

	gitRepository = "https://raw.githubusercontent.com/procinger/turing-pi-v2-cluster/refs/heads/main/"

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
