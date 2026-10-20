package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
	"github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func testGoCommand(t *testing.T) toolchaincli.GeneratedTestGoCommand {
	t.Helper()
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	command, err := toolchaincli.ObserveGeneratedTestGo(filepath.Join(runtime.GOROOT(), "bin", name), os.Environ())
	if err != nil {
		t.Fatal(err)
	}
	return command
}

func TestSelectedGoCommandUsesObservedExecutableAndEnvironment(t *testing.T) {
	tool := testGoCommand(t)
	if !filepath.IsAbs(tool.Inputs.Path) || tool.Inputs.Version != "go1.27.3" {
		t.Fatal("selected executable identity missing", tool.Inputs)
	}
	command := exec.Command(tool.Inputs.Path, "version")
	command.Env = tool.Env
	output, err := command.Output()
	if err != nil || strings.TrimSpace(string(output)) != "go version go1.27.3 "+runtime.GOOS+"/"+runtime.GOARCH {
		t.Fatal("actual selected Go differs from recorded build information", err, string(output))
	}
	if err := tool.VerifyBytes(); err != nil {
		t.Fatal(err)
	}
	tool.Env = append(tool.Env, "GOFLAGS=-race")
	if err := tool.VerifyBytes(); err == nil {
		t.Fatal("mutated command environment retained its old observation")
	}
}

func TestObserverVersionDoesNotAuthorizePATHWrapper(t *testing.T) {
	root := t.TempDir()
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	marker := filepath.Join(root, "started")
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\ntouch '"+marker+"'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root)
	if _, err := toolchaincli.CaptureGeneratedTestGo(); err == nil || !strings.Contains(err.Error(), "native cmd/go") {
		t.Fatal("observer version accepted a different child", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("tool observation executed a wrapper", err)
	}
}

func TestSelectedToolChangeClosesBeforeStartingTimedCommand(t *testing.T) {
	tool := testGoCommand(t)
	tool.Inputs.ExecutableDigest = strings.Repeat("b", 64)
	output := t.TempDir()
	if _, err := runTimed(tool, output, "test", []string{"version"}); err == nil || !strings.Contains(err.Error(), "changed after") {
		t.Fatal("changed selected tool reached execution", err)
	}
	files, err := os.ReadDir(output)
	if err != nil || len(files) != 0 {
		t.Fatal("rejected command wrote timing or execution output", err)
	}
}

func TestTimedGoCommandReadsNativeProcessResources(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("peak RSS is supported on Linux and macOS")
	}
	tool := testGoCommand(t)
	result, err := runTimed(tool, t.TempDir(), "version", []string{"version"})
	if err != nil || !result.Success || result.ExitCode != 0 || result.PeakRSSKib <= 0 || result.WallMS <= 0 || result.ResultDigest == "" {
		t.Fatal("selected child lost its actual process observations", err, result)
	}
	report := baselineReport(identity{}, result, result)
	if report.PeakRSSSubject != publictestreuse.PeakRSSGoCommands ||
		report.PeakRSSKib != result.PeakRSSKib || snapshot(report).PeakRSSSubject != report.PeakRSSSubject {
		t.Fatal("baseline summary lost its measured process", report)
	}
	if current, err := toolchaincli.CurrentPeakRSSKiB(); err != nil || current <= 1 {
		t.Fatal("current process observation is unavailable or fabricated", err, current)
	}
}

func TestTimedGoCommandPreservesLaunchAndExitFailures(t *testing.T) {
	tool := testGoCommand(t)
	_, err := runTimed(tool, filepath.Join(t.TempDir(), "missing"), "version", []string{"version"})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal("launch failure was replaced by a resource observation error", err)
	}
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("completed-process RSS requires Darwin or Linux")
	}
	result, err := runTimed(tool, t.TempDir(), "invalid", []string{"gooo-invalid-command"})
	if err != nil || result.Success || result.ExitCode == 0 || result.ResultDigest == "" {
		t.Fatal("unsuccessful child lost its exit result", result, err)
	}
}
