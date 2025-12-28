package argo

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"
)

func GetArgoApplication(applicationYaml string) (
	Application,
	error,
) {
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

func GetArgoApplicationFromGit(gitRepository string, applicationYaml string) (
	Application,
	error,
) {
	baseUrl := gitRepository + strings.TrimPrefix(applicationYaml, "../")
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

func GatherArgoAppPaths(app Application) (
	pathCollection []string,
) {
	if app.Spec.Sources != nil {
		for _, source := range app.Spec.Sources {
			if source.Path != "" {
				pathCollection = append(pathCollection, source.Path)
			}
		}
	}

	if app.Spec.Source != nil && app.Spec.Source.Path != "" {
		pathCollection = append(pathCollection, app.Spec.Source.Path)
	}

	return pathCollection
}
