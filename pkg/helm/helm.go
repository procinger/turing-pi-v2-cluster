package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"e2eutils/pkg/argo"

	"sigs.k8s.io/e2e-framework/third_party/helm"
)

type HelmArguments struct {
	Name           string
	Namespace      string
	Chart          string
	Version        string
	ValuesFilePath string
	OciRepository  string
}

func NewHelmManager(kubeConfigFile string) *helm.Manager {
	return helm.New(kubeConfigFile)
}

func AddHelmRepository(helmMgr *helm.Manager, repoURL, chartName string) error {
	return helmMgr.RunRepo(helm.WithArgs("add", "--force-update", chartName, repoURL))
}

func DeployHelmChart(helmMgr *helm.Manager, src argo.ApplicationSource, namespace string) error {
	opts, err := buildHelmArguments(src, namespace)
	if err != nil {
		return fmt.Errorf("building helm options: %w", err)
	}

	options := []helm.Option{
		helm.WithName(opts.Name),
		helm.WithNamespace(opts.Namespace),
		helm.WithChart(opts.Chart),
		helm.WithVersion(opts.Version),
		helm.WithArgs("--install", "--create-namespace"),
	}

	if opts.ValuesFilePath != "" {
		options = append(options, helm.WithArgs("-f", opts.ValuesFilePath))
	}

	if err := helmMgr.RunUpgrade(options...); err != nil {
		return fmt.Errorf("helm upgrade/install failed: %w", err)
	}

	return nil
}

func buildHelmArguments(src argo.ApplicationSource, namespace string) (*HelmArguments, error) {
	opts := &HelmArguments{
		Name:      src.Chart,
		Namespace: namespace,
		Version:   src.TargetRevision,
	}

	if strings.HasPrefix(src.RepoURL, "oci://") {
		base := strings.TrimSuffix(src.RepoURL, "/")
		opts.Chart = fmt.Sprintf("%s/%s", base, src.Chart)
		opts.OciRepository = ""
	} else {
		opts.Chart = fmt.Sprintf("%s/%s", src.Chart, src.Chart)
	}

	if src.Helm != nil && src.Helm.Values != "" {
		valuesFilePath, err := writeValuesFile(src.Chart, src.TargetRevision, src.Helm.Values)
		if err != nil {
			return nil, fmt.Errorf("writing values file: %w", err)
		}
		opts.ValuesFilePath = valuesFilePath
	}

	return opts, nil
}

func writeValuesFile(chart, revision, valuesContent string) (string, error) {
	valuesFilePath := fmt.Sprintf("values-%s-%s.yaml", chart, revision)
	filePath := filepath.Join(os.TempDir(), valuesFilePath)
	if err := os.WriteFile(filePath, []byte(valuesContent), 0o600); err != nil {
		return "", err
	}

	return filePath, nil
}
