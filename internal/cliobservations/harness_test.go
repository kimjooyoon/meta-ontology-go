// Package cliobservations exercises the actual command executable. It keeps
// native multi-case observations outside the already-large CLI unit package,
// while preserving the full public command path and race instrumentation.
package cliobservations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const exitOK = 0

var commandBinary string
var commandSourceIdentity string

type packageExecutionReceipt struct {
	Decision    string                           `json:"decision"`
	Error       string                           `json:"error,omitempty"`
	Result      *workspaceexecution.Result       `json:"result,omitempty"`
	OriginInput *bodyexecution.NativeOriginInput `json:"origin_input,omitempty"`
}

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gooo-cli-observations-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := runObservations(m, dir)
	if err := os.RemoveAll(dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

func runObservations(m *testing.M, dir string) int {
	commandBinary = filepath.Join(dir, "gooo")
	args := []string{"build", "-p=1"}
	if raceInstrumented {
		args = append(args, "-race")
	}
	if overlay := os.Getenv("GOOO_FRESH_CLI_OVERLAY"); overlay != "" {
		args = append(args, "-overlay", overlay)
	}
	args = append(args, "-o", commandBinary, "./cmd/gooo")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", args...)
	build.Dir = "../.."
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "build observed CLI:", err)
		return 1
	}
	var err error
	commandSourceIdentity, err = cliProducerSource(commandBinary)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read observed CLI build identity:", err)
		return 1
	}
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		raw, err := json.Marshal(map[string]any{"race": raceInstrumented, "command": args, "binary_removed_after_tests": true})
		if err != nil || os.WriteFile(filepath.Join(root, "cli-harness-build.json"), raw, 0600) != nil {
			return 1
		}
	}
	return m.Run()
}

func run(args []string, stdout, stderr io.Writer) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, commandBinary, args...)
	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Run(); err != nil {
		if failed, ok := err.(*exec.ExitError); ok {
			return failed.ExitCode()
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
