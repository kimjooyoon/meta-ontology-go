package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func TestPreviewOutputRejectsRepositoryInputs(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "go.mod")
	original := []byte("module fixture\n")
	if err := os.WriteFile(input, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writePreviewOutput(root, input, []byte("changed\n")); err == nil {
		t.Fatal("repository input was accepted as preview output")
	}
	got, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("repository input changed: got %q want %q", got, original)
	}
}

func TestPreviewOutputRejectsSymlinkBoundary(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside.json")
	original := []byte("original\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "preview.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := writePreviewOutput(root, link, []byte("changed\n")); err == nil || !strings.Contains(err.Error(), "outside the repository root") {
		t.Fatalf("symlink output was not rejected: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("symlink target changed: got %q want %q", got, original)
	}
}

func TestPreviewOutputRejectsExternalAliasesToRepositoryInput(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "go.mod")
	original := []byte("module fixture\n")
	if err := os.WriteFile(input, original, 0o644); err != nil {
		t.Fatal(err)
	}
	external := t.TempDir()
	hardlink := filepath.Join(external, "hardlink.json")
	if err := os.Link(input, hardlink); err != nil {
		t.Skipf("hardlinks are unavailable: %v", err)
	}
	if err := writePreviewOutput(root, hardlink, []byte("changed\n")); err == nil || !strings.Contains(err.Error(), "aliases") {
		t.Fatalf("hardlink output was not rejected: %v", err)
	}
	symlink := filepath.Join(external, "symlink.json")
	if err := os.Symlink(input, symlink); err != nil {
		t.Fatal(err)
	}
	if err := writePreviewOutput(root, symlink, []byte("changed\n")); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("external symlink output was not rejected: %v", err)
	}
	got, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("aliased repository input changed: got %q want %q", got, original)
	}
}

func TestPreviewCLIJSONRoundTrip(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "../.."))
	logical := "cmd/language-readiness-witness/predecessor-selection/pagination_test.go"
	output := filepath.Join(t.TempDir(), "pagination.json")
	command := exec.Command("go", "run", "./cmd/callback-preview", "--root", root, "--logical", logical, "--output", output)
	command.Dir = root
	if raw, err := command.CombinedOutput(); err != nil {
		t.Fatalf("callback preview CLI failed: %v\n%s", err, raw)
	}
	raw, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var preview extractor.CallbackPreviewResult
	if err := json.Unmarshal(raw, &preview); err != nil {
		t.Fatalf("callback preview JSON decode failed: %v", err)
	}
	if err := extractor.ValidateCallbackPreviewResult(preview); err != nil {
		t.Fatalf("callback preview JSON validation failed: %v", err)
	}
	encoded, err := json.Marshal(preview)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip extractor.CallbackPreviewResult
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatalf("callback preview JSON round-trip failed: %v", err)
	}
	if err := extractor.ValidateCallbackPreviewResult(roundTrip); err != nil {
		t.Fatalf("callback preview JSON round-trip validation failed: %v", err)
	}
}
