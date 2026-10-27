package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestPackageGenerateStageExportsShareTypesAndKeepBindings(t *testing.T) {
	root, fixture := t.TempDir(), t.TempDir()
	writeWorkspaceSource(t, root, "go.mod", "module stages\n\ngo 1.27.3\n")
	writeWorkspaceSource(t, root, "stage_test.go", libraryStageOracle)
	manifest := writeWorkspaceManifest(t, fixture, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"stages","activity":"Finish"},"packages":[{"path":"stages","sources":["main.gooo"]}]}`)
	for revision, source := range []string{libraryStageSource, strings.ReplaceAll(libraryStageSource, "Seed", "Declare")} {
		writeWorkspaceSource(t, fixture, "main.gooo", source)
		var out, diagnostics bytes.Buffer
		t.Run("generate", func(t *testing.T) {
			t.Setenv("PATH", t.TempDir())
			args := []string{"package", "generate", "--go-package", "stages", "--export-stage", "Plan=urn:stages:seed",
				"--export-stage", "Decide=urn:stages:finish", manifest}
			if code := run(args, &out, &diagnostics); code != exitOK {
				t.Fatal(code, diagnostics.String())
			}
		})
		for _, origin := range []string{`name="Plan" activity="urn:stages:seed"`, `name="Decide" activity="urn:stages:finish"`, `gooo:"urn:stages:value"`} {
			if !strings.Contains(out.String(), origin) {
				t.Fatal("lost semantic identity", origin)
			}
		}
		if strings.Count(out.String(), "type R_Record struct") != 1 {
			t.Fatal("stage exports duplicated record types")
		}
		writeWorkspaceSource(t, root, "generated.go", out.String())
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		command := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", ".")
		command.Dir, command.Env = root, append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
		result, err := command.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("revision %d: %v\n%s", revision, err, result)
		}
		consumer, err := os.ReadFile(filepath.Join(root, "stage_test.go"))
		if err != nil || string(consumer) != libraryStageOracle {
			t.Fatal("source rename required consumer edits", err)
		}
	}
	t.Log("two source display names; independent stage/graph outcomes=18/18; handwritten consumer edits=0")
}

func TestPackageGenerateStageExportsRejectWithoutPartialOutput(t *testing.T) {
	fixture := t.TempDir()
	writeWorkspaceSource(t, fixture, "main.gooo", libraryStageSource)
	manifest := writeWorkspaceManifest(t, fixture, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"stages","activity":"Finish"},"packages":[{"path":"stages","sources":["main.gooo"]}]}`)
	for _, exports := range [][]string{
		{"--export-stage"}, {"--export-stage", "Plan"}, {"--export-stage", "=urn:stages:seed"},
		{"--export-stage", "Plan="}, {"--export-stage", "private=urn:stages:seed"},
		{"--export-stage", "Evaluate=urn:stages:seed"}, {"--export-stage", "Plan=urn:missing"},
		{"--export-stage", "Plan=urn:stages:helper"}, {"--export-stage", "R_Record=urn:stages:seed"},
		{"--export-stage", "Plan=urn:stages:seed", "--export-stage", "Plan=urn:stages:finish"},
		{"--export-stage", "Plan=urn:stages:seed", "--export-stage", "Alias=urn:stages:seed"},
	} {
		var out, diagnostics bytes.Buffer
		args := append([]string{"package", "generate", "--go-package", "stages", manifest}, exports...)
		if code := run(args, &out, &diagnostics); code == exitOK || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatal("invalid export produced partial source", args, code, diagnostics.String())
		}
	}
	directory := t.TempDir()
	path := filepath.Join(directory, generatedFileName)
	writeWorkspaceSource(t, directory, generatedFileName, "prior bytes")
	var out, diagnostics bytes.Buffer
	args := []string{"package", "generate", "--go-package", "stages", "--out", directory, "--export-stage", "Plan=urn:missing", manifest}
	if code := run(args, &out, &diagnostics); code == exitOK {
		t.Fatal("unknown stage was accepted")
	}
	stored, err := os.ReadFile(path)
	if err != nil || string(stored) != "prior bytes" {
		t.Fatal("export failure replaced prior output", err)
	}
}
