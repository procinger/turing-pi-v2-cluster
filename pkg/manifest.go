package e2eutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

func GetKubernetesManifests(ctx context.Context, pathCollection []string) ([]k8s.Object, error) {
	if pathCollection == nil {
		return nil, errors.New("kustomization pathCollection must not be empty")
	}

	var objects []k8s.Object
	for _, source := range pathCollection {
		if source == "" {
			continue
		}

		o, err := prepareKubernetesManifests(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare manifests from source pathCollection %q: %w", source, err)
		}
		objects = append(objects, o...)
	}

	return objects, nil
}

func prepareKubernetesManifests(ctx context.Context, path string) ([]k8s.Object, error) {
	manifestPath := filepath.Join("..", path)
	manifestFS := os.DirFS(manifestPath)

	objects, err := decoder.DecodeAllFiles(ctx, manifestFS, "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML files from path %q: %w", manifestPath, err)
	}
	return objects, nil
}
