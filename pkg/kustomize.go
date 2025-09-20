package e2eutils

import (
	"bytes"
	"context"
	"e2eutils/pkg/argo"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func GetKubernetesManifests(argoApplication argo.Application) ([]k8s.Object, error) {
	var objects []k8s.Object
	var err error

	if argoApplication.Spec.Source != nil && argoApplication.Spec.Source.Path != "" {
		objects, err = prepareKubernetesManifests(*argoApplication.Spec.Source)
		if err != nil {
			return nil, err
		}

		return objects, nil
	}

	var source argo.ApplicationSource
	for _, source = range argoApplication.Spec.Sources {
		if source.Path == "" {
			continue
		}

		objects, err = prepareKubernetesManifests(source)
		if err != nil {
			return nil, err
		}
	}

	return objects, nil
}

func prepareKubernetesManifests(applicationSource argo.ApplicationSource) ([]k8s.Object, error) {
	realPath := os.DirFS("../" + applicationSource.Path)

	objects, err := decoder.DecodeAllFiles(context.TODO(), realPath, "*.yaml")
	if err != nil {
		return nil, err
	}
	return objects, nil
}

func BuildKustomization(path string) ([]*unstructured.Unstructured, error) {
	fSys := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	resMap, err := k.Run(fSys, path)
	if err != nil {
		return nil, fmt.Errorf("running kustomize: %w", err)
	}

	yamlData, err := resMap.AsYaml()
	if err != nil {
		return nil, fmt.Errorf("convert resmap to yaml: %w", err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(yamlData))

	var objects []*unstructured.Unstructured
	for {
		var o unstructured.Unstructured
		if err := dec.Decode(&o.Object); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decode yaml: %w", err)
		}

		// skip empty documents
		if len(o.Object) == 0 {
			continue
		}

		objects = append(objects, &o)
	}

	return objects, nil
}
