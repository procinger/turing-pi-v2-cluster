package helper

import (
	"fmt"
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/third_party/helm"
	"strings"
)

type HelmOptions struct {
	Name    string  `default:""`
	Namespace string `default:""`
	Chart string `default:""`
	Version string `default:""`
	Values string `default:""`
	OciRepository string `default:""`
}

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

func helmifyApp(app applicationV1Alpha1.ApplicationSource, namespace string) (HelmOptions, error) {
	fullChartName := getFullChartName(app.Chart, app.Chart)
	helmOciRepository := ""

	if strings.Contains(app.RepoURL, "oci://") {
		fullChartName = ""
		helmOciRepository = app.RepoURL
	}

	helmValues := make(map[int]string, 2)
	if app.Helm != nil {
		err := helmValuesToFile(app)
		if err != nil {
			return HelmOptions{}, err
		}

		helmValues[0] = "-f"
		helmValues[1] = "/tmp/helm-values.txt"
	}

	helmOptions := HelmOptions{
		Name: app.Chart,
		Chart: fullChartName,
		Namespace: namespace,
		Version: app.TargetRevision,
		Values: helmValues[0] + " " + helmValues[1],
		OciRepository: helmOciRepository,
	}

	return helmOptions, nil
}

func getFullChartName(helmRepoName string, helmChart string) string {
	return fmt.Sprintf("%s/%s", helmRepoName, helmChart)
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

func DeployHelmChart(helmMgr *helm.Manager, applicationSource applicationV1Alpha1.ApplicationSource, namespace string) error {
	helmOptions, err := helmifyApp(applicationSource, namespace)
	if err != nil {
		return err
	}

	err = helmMgr.RunUpgrade(
		helm.WithArgs("--install"),
		helm.WithName(helmOptions.Name),
		helm.WithNamespace(helmOptions.Namespace),
		helm.WithChart(helmOptions.Chart),
		helm.WithVersion(helmOptions.Version),
		helm.WithArgs("--create-namespace"),
		helm.WithArgs(helmOptions.Values),
		helm.WithArgs(helmOptions.OciRepository),
	)
	if err != nil {
		return err
	}

	return nil
}
