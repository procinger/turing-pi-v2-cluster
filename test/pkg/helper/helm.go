package helper

import (
	"fmt"
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/third_party/helm"
	"strings"
)

func GetHelmManager(cfg *envconf.Config) *helm.Manager {
	return helm.New(cfg.KubeconfigFile())
}

func AddHelmRepository(helmMgr *helm.Manager, helmRepoUrl string, helmChartName string) error {
	err := helmMgr.RunRepo(helm.WithArgs(
		"add --force-update",
		helmChartName,
		helmRepoUrl,
	))
	if err != nil {
		return err
	}

	return nil
}

func getFullChartName(helmRepoName string, helmChart string) string {
	return fmt.Sprintf("%s/%s", helmRepoName, helmChart)
}

func UpgradeHelmChart(helmMgr *helm.Manager, applicationSource applicationV1Alpha1.ApplicationSource, namespace string) error {
	fullChartName := applicationSource.Chart

	if ! strings.Contains(applicationSource.RepoURL, "oci://") {
		fullChartName = getFullChartName(applicationSource.Chart, applicationSource.Chart)
	}

	if applicationSource.Helm != nil {
		err := helmValuesToFile(applicationSource)
		if err != nil {
			return err
		}

		err = helmMgr.RunUpgrade(
			helm.WithName(applicationSource.Chart),
			helm.WithNamespace(namespace),
			helm.WithChart(fullChartName),
			helm.WithVersion(applicationSource.TargetRevision),
			helm.WithArgs("--install"),
			helm.WithArgs("-f", "/tmp/helm-values.txt"),
		)
		if err != nil {
			return err
		}
	} else {
		err := helmMgr.RunUpgrade(
			helm.WithName(applicationSource.Chart),
			helm.WithNamespace(namespace),
			helm.WithChart(fullChartName),
			helm.WithVersion(applicationSource.TargetRevision),
			helm.WithArgs("--install"),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func helmValuesToFile(applicationSource applicationV1Alpha1.ApplicationSource) error {
	helmValues, err := os.Create("/tmp/helm-values.txt")
	if err != nil {
		return err
	}

	_, err = helmValues.WriteString(applicationSource.Helm.Values)
	if err != nil {
		return err
	}

	err = helmValues.Close()
	if err != nil {
		return err
	}

	return nil
}

func InstallHelmChart(helmMgr *helm.Manager, applicationSource applicationV1Alpha1.ApplicationSource, namespace string) error {
	fullChartName := getFullChartName(applicationSource.Chart, applicationSource.Chart)

	helmOciRepository := ""
	if strings.Contains(applicationSource.RepoURL, "oci://") {
		fullChartName = ""
		helmOciRepository = applicationSource.RepoURL
	}

	if applicationSource.Helm != nil {
		err := helmValuesToFile(applicationSource)
		if err != nil {
			return err
		}
		err = helmMgr.RunInstall(
			helm.WithName(applicationSource.Chart),
			helm.WithNamespace(namespace),
			helm.WithChart(fullChartName),
			helm.WithVersion(applicationSource.TargetRevision),
			helm.WithArgs("--create-namespace"),
			helm.WithArgs("-f", "/tmp/helm-values.txt"),
			helm.WithArgs(helmOciRepository),
		)
		if err != nil {
			return err
		}
	} else {
		err := helmMgr.RunInstall(
			helm.WithName(applicationSource.Chart),
			helm.WithNamespace(namespace),
			helm.WithChart(fullChartName),
			helm.WithVersion(applicationSource.TargetRevision),
			helm.WithArgs("--create-namespace"),
			helm.WithArgs(helmOciRepository),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
