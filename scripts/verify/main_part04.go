package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func changedPaths(root, from, to string) ([]string, error) {
	output, err := runGit(root, "diff", "--name-only", "--diff-filter=ACMRTUXB", from+"..."+to)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}
func changedDiff(root, from, to, path string) (string, error) {
	return runGit(root, "diff", "--unified=0", from+"..."+to, "--", path)
}
func runGit(root string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}
