package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicpartialreuse"
	"github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func TestPartialReuseNativeCommandResourceObservation(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("native resident-memory observation is declared only for Darwin and Linux")
	}
	directory := t.TempDir()
	program := []byte("package generated\nfunc Ready() bool { return true }\n")
	contract := []byte("package generated\nimport \"testing\"\nfunc TestOrdersPartition(t *testing.T) { if !Ready() { t.Fatal(\"not ready\") } }\n")
	if err := preparePackage(directory, program, contract); err != nil {
		t.Fatal(err)
	}
	partition := publicpartialreuse.Partition{TestName: "TestOrdersPartition"}
	metrics, result, err := executeTests(directory, partition)
	if err != nil || !result.Success || result.ExitCode != 0 || !cache.Digest(result.ResultDigest).Known() ||
		result.ToolInputs.Validate() != nil || !filepath.IsAbs(result.ToolInputs.Path) ||
		metrics.BuildExecutions != 1 || metrics.TestExecutions != 1 || metrics.TestUnitsExecuted != 1 ||
		metrics.WallMS <= 0 || metrics.PeakRSSKib <= 0 {
		t.Fatal("native build/test observations are incomplete", metrics, result, err)
	}
	args := publicpartialreuse.TestCommandArgs(publicpartialreuse.Partition{TestName: "TestMissing"})
	if _, _, err := runTimed(directory, "unmatched", args, []string{"TestMissing"}); err == nil {
		t.Fatal("successful process with no matching test became a completed test")
	}
}

func TestPartialReuseRejectsWrapperWithoutExecutingIt(t *testing.T) {
	directory := t.TempDir()
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	marker := filepath.Join(directory, "started")
	if err := os.WriteFile(filepath.Join(directory, name), []byte("#!/bin/sh\ntouch '"+marker+"'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	if _, _, err := runTimed(directory, "wrapper", []string{"version"}, nil); err == nil {
		t.Fatal("wrapper was treated as a native Go command")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("observing the selected tool executed its wrapper", err)
	}
}

func TestPartialReuseNativeCommandPreservesLaunchAndExitFailures(t *testing.T) {
	_, _, err := runTimed(t.TempDir(), "invalid", []string{"gooo-invalid-command"}, nil)
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || !strings.Contains(err.Error(), "gooo-invalid-command") {
		t.Fatal("native exit error was replaced by a resource error", err)
	}
	t.Setenv("PATH", t.TempDir())
	if _, _, err := runTimed(t.TempDir(), "missing", []string{"version"}, nil); !errors.Is(err, exec.ErrNotFound) {
		t.Fatal("launch failure was replaced by a resource error", err)
	}
	if _, err := toolchaincli.PeakRSSKiB(nil); err == nil {
		t.Fatal("missing process state invented a resource observation")
	}
}
