package e2eutils

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func GetClient() klient.Client {
	cfg := envconf.Config{}
	return cfg.Client()
}

func GetRestConfig() *rest.Config {
	client := GetClient()
	return client.RESTConfig()
}

func GetClientSet() (*kubernetes.Clientset, error) {
	kubeConfig := GetRestConfig()
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func GetDynClient() (*dynamic.DynamicClient, error) {
	kubeConfig := GetRestConfig()
	return dynamic.NewForConfig(kubeConfig)
}

func GetDiscoveryClient() (*discovery.DiscoveryClient, error) {
	kubeConfig := GetRestConfig()
	return discovery.NewDiscoveryClientForConfig(kubeConfig)
}
