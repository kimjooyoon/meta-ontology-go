package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

func gitOutput(root string, args ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return output, nil
}

func verifyGitIdentity(root, expected string) error {
	if expected == "" {
		return fmt.Errorf("expected SHA is required")
	}
	head, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	actual := strings.TrimSpace(string(head))
	if actual != expected {
		return fmt.Errorf("HEAD %s does not match expected %s", actual, expected)
	}
	status, err := gitOutput(root, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return err
	}
	if len(status) != 0 {
		return fmt.Errorf("projection source must be a clean exact checkout")
	}
	return nil
}

func trackedPaths(root string) ([]string, error) {
	output, err := gitOutput(root, "ls-files", "-z")
	if err != nil {
		return nil, err
	}
	fields := bytes.Split(output, []byte{0})
	paths := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) != 0 {
			paths = append(paths, string(field))
		}
	}
	sort.Strings(paths)
	return paths, nil
}
