package argo

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
)

func GetArgoApplication(applicationYaml string) (Application, error) {
	yamlFile, err := os.ReadFile(applicationYaml)
	if err != nil {
		slog.Error("Failed to open application yaml file " + applicationYaml)
		return Application{}, err
	}

	argoApplication := &Application{}
	err = yaml.Unmarshal([]byte(yamlFile), &argoApplication)
	if err != nil {
		slog.Error("Failed to unmarshal argo app")
		return Application{}, err
	}

	return *argoApplication, nil
}

func GetArgoApplicationFromGit(applicationYaml string) (Application, error) {
	baseUrl := "https://raw.githubusercontent.com/procinger/turing-pi-v2-cluster/refs/heads/main/" + strings.TrimPrefix(applicationYaml, "../")
	response, err := http.Get(baseUrl)
	if err != nil {
		slog.Warn("Failed to fetch application yaml (" + baseUrl + ") from git")
		return Application{}, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		slog.Warn("Failed to fetch application yaml (" + baseUrl + ") from git. Server gave " + strconv.Itoa(response.StatusCode))
		return Application{}, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Application{}, err
	}

	argoApplication := &Application{}
	err = yaml.Unmarshal([]byte(body), &argoApplication)
	if err != nil {
		slog.Error("Failed to unmarshal argo application")
		return Application{}, err
	}

	return *argoApplication, nil
}
