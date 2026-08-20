package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func inspectGit(settings cutoverConfig) (gitState, error) {
	state := gitState{}
	head, err := cutoverGit(settings.root, nil, "rev-parse", "HEAD")
	if err != nil {
		return state, err
	}
	state.Head = strings.TrimSpace(string(head))
	status, err := cutoverGit(settings.root, nil, "status", "--porcelain")
	if err != nil {
		return state, err
	}
	if len(bytes.TrimSpace(status)) != 0 {
		state.Dirty = 1
	}
	tracked, err := cutoverGit(settings.root, nil, "ls-files", "-z")
	if err != nil {
		return state, err
	}
	state.Tracked = zeroStrings(tracked)
	return state, nil
}

func cutoverGit(root string, input []byte, arguments ...string) ([]byte, error) {
	command := exec.Command("git", arguments...)
	command.Dir = root
	command.Env = os.Environ()
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %s: %w", strings.Join(arguments, " "), output, err)
	}
	return output, nil
}

func zeroStrings(data []byte) []string {
	values := []string{}
	for _, value := range bytes.Split(data, []byte{0}) {
		if len(value) != 0 {
			values = append(values, string(value))
		}
	}
	return values
}
