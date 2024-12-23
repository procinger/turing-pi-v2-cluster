package test

import (
	"fmt"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
	"test/test/pkg/types/argocd"
	"testing"
)

var (
	ciTestEnv       env.Environment
	kindClusterName string
	argoAppCurrent  argocd.Application
	argoAppUpdate   argocd.Application
)

func TestMain(m *testing.M) {
	config, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Println("Could not create config from env", err)
	}

	ciTestEnv = env.NewWithConfig(config)
	kindClusterName = envconf.RandomName("ci-e2e-test", 16)

	ciTestEnv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
	)

	ciTestEnv.Finish(
		envfuncs.ExportClusterLogs(kindClusterName, "./kind-logs"),
		envfuncs.DestroyCluster(kindClusterName),
	)

	os.Exit(ciTestEnv.Run(m))
}
