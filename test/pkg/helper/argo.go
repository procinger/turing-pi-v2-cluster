package helper

import (
	"fmt"
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"io"
	"net/http"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
)

func GetArgoApplication(applicationYaml string) (applicationV1Alpha1.Application, error) {
	yamlFile, err := os.ReadFile(applicationYaml)
	if err != nil {
		fmt.Printf("Failed to open application yaml file. Error #%v ", err)
		return applicationV1Alpha1.Application{}, err
	}

	argoApplication := &applicationV1Alpha1.Application{}
	err = yaml.Unmarshal([]byte(yamlFile), &argoApplication)
	if err != nil {
		fmt.Printf("Failed to unmarshal argo app #%v", err)
		return applicationV1Alpha1.Application{}, err
	}

	return *argoApplication, nil
}

func GetArgoApplicationFromGit(applicationYaml string) (applicationV1Alpha1.Application, error) {
	baseUrl := "https://raw.githubusercontent.com/procinger/turing-pi-v2-cluster/refs/heads/main/" + strings.TrimPrefix(applicationYaml, "../")
	response, err := http.Get(baseUrl)
	if err != nil {
		fmt.Printf("Failed to fetch application yaml from git. Error #%v ", err)
		return applicationV1Alpha1.Application{}, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch application yaml from git. Server gave #%v ", response.StatusCode)
		return applicationV1Alpha1.Application{}, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return applicationV1Alpha1.Application{}, err
	}

	argoApplication := &applicationV1Alpha1.Application{}
	err = yaml.Unmarshal([]byte(body), &argoApplication)
	if err != nil {
		fmt.Printf("Failed to unmarshal argo app #%v", err)
		return applicationV1Alpha1.Application{}, err
	}

	return *argoApplication, nil
}
