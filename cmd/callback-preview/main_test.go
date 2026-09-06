package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if err := writePreviewOutput(root, link, []byte("changed\n")); err == nil || !strings.Contains(err.Error(), "symlink") {
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
