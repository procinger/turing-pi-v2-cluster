package test

import (
	"fmt"
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"testing"
)

var (
	ciTestEnv       env.Environment
	kindClusterName string
	argoAppCurrent  applicationV1Alpha1.Application
	argoAppUpdate   applicationV1Alpha1.Application
)

func TestMain(m *testing.M) {
	config, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Println("Could not create config from env", err)
	}

	ciTestEnv = env.NewWithConfig(config)
	kindClusterName = envconf.RandomName("ci-e2e-test", 16)

	ciTestEnv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
	)

	ciTestEnv.Finish(
		envfuncs.ExportKindClusterLogs(kindClusterName, "./kind-logs"),
		envfuncs.DestroyKindCluster(kindClusterName),
	)

	os.Exit(ciTestEnv.Run(m))
}
