package main

import (
	"os"
	"path/filepath"
	"testing"
)

// CLI-PATH-001: output roots are absolute and canonical before use.
func TestCLIPATH001CanonicalOutputRoot(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := canonicalOutputRoot(filepath.Join(root, "nested", "..", "nested"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("canonical root = %q, want %q", got, want)
	}
}

// CLI-PATH-002 and CLI-PATH-003: output names cannot traverse or be absolute.
func TestCLIPATH002003RejectEscapingOutputNames(t *testing.T) {
	root := t.TempDir()
	if _, err := resolveOutputPath(root, filepath.Join("..", "escape.go")); err == nil {
		t.Fatal("traversal output name was accepted")
	}
	if _, err := resolveOutputPath(root, filepath.Join(root, "escape.go")); err == nil {
		t.Fatal("absolute output name was accepted")
	}
}

// CLI-PATH-004: a symlink cannot become the trusted output root.
func TestCLIPATH004RejectsSymlinkOutputRoot(t *testing.T) {
	actual := t.TempDir()
	parent := t.TempDir()
	link := filepath.Join(parent, "out")
	if err := os.Symlink(actual, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := canonicalOutputRoot(link); err == nil {
		t.Fatal("symlink output root was accepted")
	}
}
