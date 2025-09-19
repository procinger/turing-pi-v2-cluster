package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"e2eutils/pkg/argo"

	"sigs.k8s.io/e2e-framework/third_party/helm"
)

func NewHelmManager(kubeConfigFile string) *helm.Manager {
	return helm.New(kubeConfigFile)
}

func AddHelmRepository(helmMgr *helm.Manager, repoURL, chartName string) error {
	return helmMgr.RunRepo(helm.WithArgs("add", "--force-update", chartName, repoURL))
}

func DeployHelmChart(helmMgr *helm.Manager, src argo.ApplicationSource, namespace string) error {
	opts, err := buildHelmOptions(src, namespace)
	if err != nil {
		return fmt.Errorf("building helm options: %w", err)
	}

	if err := helmMgr.RunUpgrade(opts...); err != nil {
		return fmt.Errorf("helm upgrade/install failed: %w", err)
	}

	return nil
}

func buildHelmOptions(src argo.ApplicationSource, namespace string) ([]helm.Option, error) {
	chart := ""
	if strings.HasPrefix(src.RepoURL, "oci://") {
		base := strings.TrimSuffix(src.RepoURL, "/")
		chart = fmt.Sprintf("%s/%s", base, src.Chart)
	} else {
		chart = fmt.Sprintf("%s/%s", src.Chart, src.Chart)
	}

	options := []helm.Option{
		helm.WithName(src.Chart),
		helm.WithNamespace(namespace),
		helm.WithChart(chart),
		helm.WithVersion(src.TargetRevision),
		helm.WithArgs("--install", "--create-namespace"),
	}

	if src.Helm != nil && src.Helm.Values != "" {
		valuesFilePath, err := writeValuesFile(src.Chart, src.TargetRevision, src.Helm.Values)
		if err != nil {
			return nil, fmt.Errorf("writing values file: %w", err)
		}

		options = append(options, helm.WithArgs("-f", valuesFilePath))
	}

	return options, nil
}

func writeValuesFile(chart, revision, valuesContent string) (string, error) {
	valuesFilePath := fmt.Sprintf("values-%s-%s.yaml", chart, revision)
	filePath := filepath.Join(os.TempDir(), valuesFilePath)
	if err := os.WriteFile(filePath, []byte(valuesContent), 0o600); err != nil {
		return "", err
	}

	return filePath, nil
}
