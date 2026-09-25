package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CLI-PATH-005: rejecting an output symlink preserves its target metadata.
func TestCLIPATH005RejectsOutputSymlinkWithoutMutation(t *testing.T) {
	outputDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside.go")
	original := []byte("outside bytes\n")
	if err := os.WriteFile(target, original, 0o640); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(outputDir, generatedFileName)
	if err := os.Symlink(target, output); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	code := runGenerate([]string{"billing.gooo", "--out", outputDir}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != exitFailure {
		t.Fatalf("symlink output exit code = %d, want %d", code, exitFailure)
	}
	after, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	linkTarget, err := os.Readlink(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) || linkTarget != target {
		t.Fatal("rejected symlink output changed bytes, inode, mode, mtime, or target")
	}
}

// CLI-PATH-006: input symlinks are rejected before parsing.
func TestCLIPATH006RejectsInputSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "source.gooo")
	link := filepath.Join(dir, "link.gooo")
	if err := os.WriteFile(target, []byte(validSource), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	var stderr bytes.Buffer
	code := runCheck([]string{link}, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitFailure || !strings.Contains(stderr.String(), "symbolic link") {
		t.Fatalf("symlink input = code %d, stderr %q", code, stderr.String())
	}
}
