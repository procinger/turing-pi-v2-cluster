package helper

import (
	"fmt"
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"path"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

func GetHelmManager(cfg *envconf.Config) *helm.Manager {
	return helm.New(cfg.KubeconfigFile())
}

func AddHelmRepository(helmMgr *helm.Manager, helmRepoUrl string) error {
	helmRepoName := path.Base(helmRepoUrl)
	err := helmMgr.RunRepo(helm.WithArgs(
		"add",
		helmRepoName,
		helmRepoUrl,
	))
	if err != nil {
		return err
	}

	return nil
}

func getHelmRepoName(helmRepoUrl string) string {
	return path.Base(helmRepoUrl)
}

func getFullChartName(helmRepoName string, helmChart string) string {
	return fmt.Sprintf("%s/%s", helmRepoName, helmChart)
}

func UpgradeHelmChart(helmMgr *helm.Manager, applicationSource *applicationV1Alpha1.ApplicationSource, namespace string) error {
	helmRepoUrl := applicationSource.RepoURL
	helmRepoName := getHelmRepoName(helmRepoUrl)

	fullChartName := getFullChartName(helmRepoName, applicationSource.Chart)
	if applicationSource.Helm != nil {
		err := helmMgr.RunUpgrade(
			helm.WithName(applicationSource.Chart),
			helm.WithNamespace(namespace),
			helm.WithChart(fullChartName),
			helm.WithVersion(applicationSource.TargetRevision),
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
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func InstallHelmChart(helmMgr *helm.Manager, applicationSource *applicationV1Alpha1.ApplicationSource, namespace string) error {
	helmRepoUrl := applicationSource.RepoURL
	helmRepoName := getHelmRepoName(helmRepoUrl)

	fullChartName := getFullChartName(helmRepoName, applicationSource.Chart)
	if applicationSource.Helm != nil {
		err := helmMgr.RunInstall(
			helm.WithName(applicationSource.Chart),
			helm.WithNamespace(namespace),
			helm.WithChart(fullChartName),
			helm.WithVersion(applicationSource.TargetRevision),
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
		)
		if err != nil {
			return err
		}
	}

	return nil
}
