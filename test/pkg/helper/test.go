package helper

import (
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

func PrepareTest(applicationYaml string, argoAppCurrent *applicationV1Alpha1.Application, argoAppUpdate *applicationV1Alpha1.Application) error {
	currGitBranch, err := GetCurrentGitBranch()
	if err != nil {
		return err
	}

	if currGitBranch != "main" {
		*argoAppUpdate, err = GetArgoApplication(applicationYaml)
		err = CheckoutGitBranch("main")
		if err != nil {
			return err
		}
		*argoAppCurrent, err = GetArgoApplication(applicationYaml)
		if err != nil {
			return err
		}
	} else {
		*argoAppCurrent, err = GetArgoApplication("../kubernetes-services/templates/sealed-secrets.yaml")
		if err != nil {
			return err
		}
	}

	return nil
}
