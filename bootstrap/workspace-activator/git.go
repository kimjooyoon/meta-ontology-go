package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func inspectGit(settings activationConfig) (int, int, error) {
	status, err := activationGit(settings, "status", "--porcelain")
	if err != nil {
		return 1, 1, err
	}
	head, err := activationGit(settings, "rev-parse", "HEAD")
	if err != nil {
		return 1, 1, err
	}
	dirty := 0
	if strings.TrimSpace(status) != "" {
		dirty = 1
	}
	drift := 0
	if strings.TrimSpace(head) != settings.expectedSHA {
		drift = 1
	}
	return dirty, drift, nil
}

func activationGit(settings activationConfig, arguments ...string) (string, error) {
	command := exec.Command("git", arguments...)
	command.Dir = settings.root
	command.Env = append(os.Environ(),
		"GIT_DIR="+settings.gitDir,
		"GIT_WORK_TREE="+settings.root,
		"GIT_INDEX_FILE="+settings.gitIndex,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(arguments, " "), output, err)
	}
	return string(output), nil
}
