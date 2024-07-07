package helper

import (
	"fmt"
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"io/ioutil"
	"sigs.k8s.io/yaml"
)

func GetArgoApplication(applicationYaml string) (applicationV1Alpha1.Application, error) {
	yamlFile, err := ioutil.ReadFile(applicationYaml)
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
