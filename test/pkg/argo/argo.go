package argo

import (
	"errors"
	"io"
	"net/http"
	"os"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
)

func GetArgoApplication(applicationYaml string) (Application, error) {
	yamlFile, err := os.ReadFile(applicationYaml)
	if err != nil {
		return Application{}, errors.New("Failed to open application yaml file " + applicationYaml + ". " + err.Error())
	}

	argoApplication := &Application{}
	err = yaml.Unmarshal([]byte(yamlFile), &argoApplication)
	if err != nil {
		return Application{}, errors.New("Failed to unmarshal argo app." + err.Error())
	}

	return *argoApplication, nil
}

func GetArgoApplicationFromGit(applicationYaml string) (Application, error) {
	baseUrl := "https://raw.githubusercontent.com/procinger/turing-pi-v2-cluster/refs/heads/main/" + strings.TrimPrefix(applicationYaml, "../")
	response, err := http.Get(baseUrl)
	if err != nil {
		return Application{}, errors.New("Failed to fetch application yaml (" + baseUrl + ") from git. " + err.Error())
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Application{}, errors.New("Failed to fetch application yaml (" + baseUrl + ") from git. Server gave " + strconv.Itoa(response.StatusCode))
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Application{}, err
	}

	argoApplication := &Application{}
	err = yaml.Unmarshal([]byte(body), &argoApplication)
	if err != nil {
		return Application{}, errors.New("Failed to unmarshal argo app. " + err.Error())
	}

	return *argoApplication, nil
}
