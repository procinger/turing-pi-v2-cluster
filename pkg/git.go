package e2eutils

import (
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v6"
)

func Clone(path string, gitRepository string) error {
	repo, err := git.PlainClone(path, &git.CloneOptions{
		URL:      gitRepository,
		Depth:    1,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	commit, err := repo.Head()
	if err != nil {
		return err
	}

	println(commit.Hash().String())
	return nil
}

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
