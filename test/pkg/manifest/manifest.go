package manifest

import (
	"context"
	"os"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"strings"
	"test/test/pkg/argo"
)

func GetKubernetesManifests(argoApplication argo.Application) ([]k8s.Object, error) {
	var objects []k8s.Object
	var err error

	if argoApplication.Spec.Source != nil {
		if argoApplication.Spec.Source.Path == "" {
			return nil, nil
		}

		objects, err = prepareKubernetesManifests(*argoApplication.Spec.Source)
		if err != nil {
			return nil, err
		}
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

func BuildKustomization(path string) ([]string, error) {
	fSys := filesys.MakeFsOnDisk()
	kustomizationDir := "../" + path
	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	objects, err := k.Run(fSys, kustomizationDir)
	if err != nil {
		return nil, err
	}

	yaml, err := objects.AsYaml()
	if err != nil {
		return nil, err
	}

	return strings.Split(string(yaml), "---"), nil
}
