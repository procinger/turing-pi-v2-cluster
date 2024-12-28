package helper

import (
	"bytes"
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"log/slog"
	"os"
	"path/filepath"
	"test/test/pkg/types/argocd"
)

func GetKubernetesManifests(argoApplication argocd.Application) ([]runtime.Object, error) {
	var yamlCollection []byte
	var err error

	if argoApplication.Spec.Source != nil {
		if argoApplication.Spec.Source.Path == "" {
			return nil, nil
		}

		yamlCollection, err = prepareKubernetesManifests(*argoApplication.Spec.Source)
		if err != nil {
			return nil, err
		}
	}

	var source argocd.ApplicationSource
	for _, source = range argoApplication.Spec.Sources {
		if source.Path == "" {
			continue
		}

		yamlCollection, err = prepareKubernetesManifests(source)
		if err != nil {
			return nil, err
		}
	}

	objects, err := unmarshal(yamlCollection)
	if err != nil {
		return nil, err
	}
	return objects, nil
}

func prepareKubernetesManifests(applicationSource argocd.ApplicationSource) ([]byte, error) {
	realPath := "../" + applicationSource.Path
	yamlFiles, err := os.ReadDir(realPath)
	if err != nil {
		slog.Error("Failed to read directory " + realPath)
		return nil, err
	}

	var yamlCollection []byte

	for _, file := range yamlFiles {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		filePath := filepath.Join(realPath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		yamlCollection = append(yamlCollection, "---\n"...)
		yamlCollection = append(yamlCollection, data...)
	}

	return yamlCollection, nil
}

func unmarshal(yaml []byte) ([]runtime.Object, error) {
	var objectList []runtime.Object
	yamlFiles := bytes.Split(yaml, []byte("---"))
	for _, file := range yamlFiles {
		if len(file) == 0 || string(file) == "\n" {
			continue
		}

		object, err := Decode(file)
		if err != nil {
			slog.Error("Failed to decode kubernetes resource " + string(file))
			return nil, err
		}

		objectList = append(objectList, object)
	}

	return objectList, nil
}

func Decode(data []byte) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	return obj, err
}

func Apply(clientset kubernetes.Clientset, object runtime.Object) error {
	switch object.(type) {
	case *corev1.Namespace:
		_, err := createNamespace(clientset, *object.(*corev1.Namespace))
		return err
	case *corev1.Secret:
		_, err := createSecret(clientset, *object.(*corev1.Secret))
		return err
	case *appsv1.Deployment:
		_, err := createDeployment(clientset, *object.(*appsv1.Deployment))
		return err
	default:

	}

	return nil
}

func createNamespace(clientset kubernetes.Clientset, object corev1.Namespace) (*corev1.Namespace, error) {
	return clientset.CoreV1().Namespaces().Create(context.TODO(), &object, metav1.CreateOptions{})
}

func createSecret(clientset kubernetes.Clientset, object corev1.Secret) (*corev1.Secret, error) {
	return clientset.CoreV1().Secrets(object.Namespace).Create(context.TODO(), &object, metav1.CreateOptions{})
}

func createDeployment(clientset kubernetes.Clientset, object appsv1.Deployment) (*appsv1.Deployment, error) {
	return clientset.AppsV1().Deployments(object.Namespace).Create(context.TODO(), &object, metav1.CreateOptions{})
}
