package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type registrationProcess struct {
	observation generation.ProcessObservation
	stdout      []byte
	stderr      []byte
	err         error
}

// Each observation is derived from a real child process. Canonical commands
// identify argument roles; raw output hashes preserve the actual execution.
func registrationRun(root string, descriptor []string, program string, args ...string) registrationProcess {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, program, args...)
	command.Dir = root
	command.Env = registrationEnvironment(root)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	code := 0
	if err != nil {
		code = -1
		if command.ProcessState != nil {
			code = command.ProcessState.ExitCode()
		}
	}
	out, diagnostic := stdout.Bytes(), stderr.Bytes()
	return registrationProcess{observation: descriptorObservation(descriptor, out, diagnostic, code),
		stdout: out, stderr: diagnostic, err: err}
}

func registrationEnvironment(root string) []string {
	var result []string
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "GIT_") || strings.HasPrefix(value, "GOWORK=") ||
			strings.HasPrefix(value, "GOTOOLCHAIN=") || strings.HasPrefix(value, "LOGICAL_WORKSPACE=") ||
			strings.HasPrefix(value, "GITHUB_WORKSPACE=") {
			continue
		}
		result = append(result, value)
	}
	return append(result, "GOWORK=off", "GOTOOLCHAIN=local",
		"LOGICAL_WORKSPACE="+root, "GITHUB_WORKSPACE="+root)
}

func registrationProcessFailure(step string, process registrationProcess) *operationError {
	failure := newOperationError("SYNTAX_REGISTRATION", step, "REGISTRATION_PROCESS_FAILED",
		"DIRECT_MISSING", "inspect-process-diagnostic-and-retry")
	failure.diagnostics = []string{boundedCollapseDiagnostic(process.stdout), boundedCollapseDiagnostic(process.stderr)}
	return failure
}
