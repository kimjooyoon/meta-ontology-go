package externalcapabilityexecution

import (
	"fmt"
	"os"
	"path/filepath"
)

type capabilityTools struct {
	Evaluator string
	Gomacro   string
}

func prepareTools(workspace, externalRoot string) (capabilityTools, error) {
	harness := filepath.Join(workspace, "evaluator")
	if err := os.MkdirAll(harness, 0o755); err != nil {
		return capabilityTools{}, err
	}
	goMod := fmt.Sprintf("module gooo.external.capability.witness\n\ngo 1.27.0\n\n"+
		"require github.com/cosmos72/gomacro v0.0.0\n\n"+
		"replace github.com/cosmos72/gomacro => %s\n", filepath.ToSlash(externalRoot))
	if err := os.WriteFile(filepath.Join(harness, "go.mod"), []byte(goMod), 0o644); err != nil {
		return capabilityTools{}, err
	}
	if err := os.WriteFile(filepath.Join(harness, "main.go"), []byte(evaluatorSource), 0o644); err != nil {
		return capabilityTools{}, err
	}
	tools := capabilityTools{
		Evaluator: filepath.Join(workspace, "evaluate-gomacro"),
		Gomacro:   filepath.Join(workspace, "gomacro"),
	}
	code, output, err := runCommand(harness, commandEnvironment("GOWORK=off", "GOFLAGS=-mod=mod"),
		"go", "build", "-o", tools.Evaluator, ".")
	if err != nil || code != 0 {
		return capabilityTools{}, fmt.Errorf("build evaluator (%d): %w: %s", code, err, output)
	}
	code, output, err = runCommand(externalRoot, commandEnvironment("GOWORK=off", "GOFLAGS=-mod=readonly"),
		"go", "build", "-o", tools.Gomacro, ".")
	if err != nil || code != 0 {
		return capabilityTools{}, fmt.Errorf("build gomacro (%d): %w: %s", code, err, output)
	}
	return tools, nil
}
