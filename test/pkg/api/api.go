package api

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"log/slog"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"test/test/pkg/test"
)

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
	case *corev1.PersistentVolumeClaim:
		_, err := createPersistentVolumeClaim(clientset, *object.(*corev1.PersistentVolumeClaim))
		return err
	case *unstructured.Unstructured:
		dynClient, err := test.GetDynClient()
		if err != nil {
			return err
		}
		err = createCustomResourceDefinition(dynClient, *object.(*unstructured.Unstructured))
		return err
	default:

	}

	return nil
}

func ApplyAll(clientset kubernetes.Clientset, objectList []k8s.Object) error {
	for _, object := range objectList {
		err := Apply(clientset, object)
		if err != nil {
			return err
		}
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

func createPersistentVolumeClaim(clientset kubernetes.Clientset, object corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	return clientset.CoreV1().PersistentVolumeClaims(object.Namespace).Create(context.TODO(), &object, metav1.CreateOptions{})
}

func getResourceName(object unstructured.Unstructured) (string, error) {
	discoveryClient, err := test.GetDiscoveryClient()
	if err != nil {
		slog.Error("Failed to get API resources: " + err.Error())
		return "", err
	}

	apiResources, err := discoveryClient.ServerResourcesForGroupVersion(object.GetAPIVersion())
	if err != nil {
		slog.Error("Failed to get API resources: " + err.Error())
		return "", err
	}

	var resourceName string
	for _, resource := range apiResources.APIResources {
		if resource.Kind == object.GetKind() {
			resourceName = resource.Name
			break
		}
	}

	return resourceName, nil
}

func createCustomResourceDefinition(dynClient *dynamic.DynamicClient, object unstructured.Unstructured) error {
	resourceName, err := getResourceName(object)
	if err != nil {
		slog.Error("Failed to create resource: " + err.Error())
	}
	
	gvr := schema.GroupVersionResource{
		Group:    object.GroupVersionKind().Group,
		Version:  object.GroupVersionKind().Version,
		Resource: resourceName,
	}

	_, err = dynClient.Resource(gvr).Namespace(object.GetNamespace()).Create(context.TODO(), &object, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create resource: " + err.Error())
	}

	return nil
}
