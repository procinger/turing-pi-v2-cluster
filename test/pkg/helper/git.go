package helper

import (
	"os/exec"
	"strings"
)

func GetCurrentGitBranch() (string, error) {
	cmd := exec.Command("git", []string{"branch", "--show-current"}...)
	currGitBranch, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(currGitBranch), "\n"), nil
}

func CheckoutGitBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
