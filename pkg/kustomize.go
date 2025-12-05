package e2eutils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func BuildKustomization(path string) ([]*unstructured.Unstructured, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("kustomization path must not be empty")
	}

	fs := filesys.MakeFsOnDisk()
	kustomizer := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	resMap, err := kustomizer.Run(fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to run kustomize: %w", err)
	}

	yamlData, err := resMap.AsYaml()
	if err != nil {
		return nil, fmt.Errorf("failed to convert resource map to YAML: %w", err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(yamlData))

	var objects []*unstructured.Unstructured
	for {
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj.Object); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}

		// skip empty documents
		if len(obj.Object) == 0 {
			continue
		}

		objects = append(objects, &obj)
	}

	return objects, nil
}
