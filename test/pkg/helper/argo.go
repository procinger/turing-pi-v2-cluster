package helper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"test/test/pkg/types/argocd"
)

func GetArgoApplication(applicationYaml string) (argocd.Application, error) {
	yamlFile, err := os.ReadFile(applicationYaml)
	if err != nil {
		fmt.Printf("Failed to open application yaml file. Error #%v ", err)
		return argocd.Application{}, err
	}

	argoApplication := &argocd.Application{}
	err = yaml.Unmarshal([]byte(yamlFile), &argoApplication)
	if err != nil {
		fmt.Printf("Failed to unmarshal argo app #%v", err)
		return argocd.Application{}, err
	}

	return *argoApplication, nil
}

func GetArgoApplicationFromGit(applicationYaml string) (argocd.Application, error) {
	baseUrl := "https://raw.githubusercontent.com/procinger/turing-pi-v2-cluster/refs/heads/main/" + strings.TrimPrefix(applicationYaml, "../")
	response, err := http.Get(baseUrl)
	if err != nil {
		fmt.Printf("Failed to fetch application yaml from git. Error #%v ", err)
		return argocd.Application{}, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch application yaml from git. Server gave #%v ", response.StatusCode)
		return argocd.Application{}, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return argocd.Application{}, err
	}

	argoApplication := &argocd.Application{}
	err = yaml.Unmarshal([]byte(body), &argoApplication)
	if err != nil {
		fmt.Printf("Failed to unmarshal argo app #%v", err)
		return argocd.Application{}, err
	}

	return *argoApplication, nil
}
