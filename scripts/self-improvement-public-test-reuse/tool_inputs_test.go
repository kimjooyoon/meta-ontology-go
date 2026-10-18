package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
	"github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func testGoCommand(t *testing.T) goCommand {
	t.Helper()
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	command, err := observeGoCommand(filepath.Join(runtime.GOROOT(), "bin", name), os.Environ())
	if err != nil {
		t.Fatal(err)
	}
	return command
}

func TestSelectedGoCommandUsesObservedExecutableAndEnvironment(t *testing.T) {
	tool := testGoCommand(t)
	if !filepath.IsAbs(tool.Inputs.Path) || tool.Inputs.Version != "go1.27.2" {
		t.Fatal("selected executable identity missing", tool.Inputs)
	}
	command := exec.Command(tool.Inputs.Path, "version")
	command.Env = tool.Env
	output, err := command.Output()
	if err != nil || strings.TrimSpace(string(output)) != "go version go1.27.2 "+runtime.GOOS+"/"+runtime.GOARCH {
		t.Fatal("actual selected Go differs from recorded build information", err, string(output))
	}
	if err := tool.verifyBytes(); err != nil {
		t.Fatal(err)
	}
	tool.Env = append(tool.Env, "GOFLAGS=-race")
	if err := tool.verifyBytes(); err == nil {
		t.Fatal("mutated command environment retained its old observation")
	}
}

func TestGoCommandEnvironmentIsFrozenAndCanonical(t *testing.T) {
	a := goCommandEnvironment([]string{"HOME=/home/a", "GOCACHE=/cache", "PATH=/bin", "GOFLAGS=-race", "GOENV=/custom", "TOKEN=private"}, "/sdk")
	b := goCommandEnvironment([]string{"TOKEN=other", "PATH=/bin", "GOCACHE=/cache", "HOME=/home/a", "GOFLAGS=-tags=other"}, "/sdk")
	if !slices.Equal(a, b) || !slices.IsSorted(a) {
		t.Fatal("ambient Go settings or credentials changed the selected profile")
	}
	for _, entry := range []string{"GOENV=off", "GOFLAGS=-mod=readonly", "GOTOOLCHAIN=local", "GOWORK=off", "GOROOT=/sdk", "GOPROXY=off", "CGO_ENABLED=0"} {
		if !slices.Contains(a, entry) {
			t.Fatal("missing controlled setting", entry)
		}
	}
	c := goCommandEnvironment([]string{"PATH=/other", "HOME=/home/a", "GOCACHE=/cache"}, "/sdk")
	if slices.Equal(a, c) {
		t.Fatal("actual retained environment change was discarded")
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
	if _, err := captureGoCommand(); err == nil || !strings.Contains(err.Error(), "native cmd/go") {
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
