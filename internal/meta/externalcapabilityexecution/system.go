package externalcapabilityexecution

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runCommand(directory string, environment []string, name string, arguments ...string) (int, []byte, error) {
	command := exec.Command(name, arguments...)
	command.Dir = directory
	if environment != nil {
		command.Env = environment
	}
	output, err := command.CombinedOutput()
	if err == nil {
		return 0, output, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode(), output, nil
	}
	return -1, output, err
}

func runText(directory string, name string, arguments ...string) (string, error) {
	code, output, err := runCommand(directory, nil, name, arguments...)
	if err != nil || code != 0 {
		return "", fmt.Errorf("%s failed with %d: %w", name, code, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func commandEnvironment(overrides ...string) []string {
	environment := os.Environ()
	for _, override := range overrides {
		key := strings.SplitN(override, "=", 2)[0] + "="
		filtered := environment[:0]
		for _, item := range environment {
			if !strings.HasPrefix(item, key) {
				filtered = append(filtered, item)
			}
		}
		environment = append(filtered, override)
	}
	return environment
}
