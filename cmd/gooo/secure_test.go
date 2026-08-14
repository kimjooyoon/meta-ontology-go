package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

func TestCLIPATH004CommandRejectsSymlinkRootWithoutMutation(t *testing.T) {
	actual := t.TempDir()
	parent := t.TempDir()
	link := filepath.Join(parent, "out")
	if err := os.Symlink(actual, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	sentinel := filepath.Join(actual, "sentinel.go")
	original := []byte("sentinel bytes\n")
	if err := os.WriteFile(sentinel, original, 0o640); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(sentinel)
	if err != nil {
		t.Fatal(err)
	}
	beforeActual := directoryEntries(t, actual)
	beforeParent := directoryEntries(t, parent)
	run := func() (int, string, string) {
		var stdout, stderr bytes.Buffer
		code := runGenerate([]string{"billing.gooo", "--out", link}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
	firstCode, firstOut, firstErr := run()
	secondCode, secondOut, secondErr := run()
	if firstCode != exitFailure || secondCode != exitFailure || firstOut != "" || secondOut != "" || firstErr != secondErr || !strings.Contains(firstErr, "symbolic link") {
		t.Fatalf("symlink root rejection changed: first=%d/%q/%q second=%d/%q/%q", firstCode, firstOut, firstErr, secondCode, secondOut, secondErr)
	}
	after, err := os.Stat(sentinel)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) || !reflect.DeepEqual(beforeActual, directoryEntries(t, actual)) || !reflect.DeepEqual(beforeParent, directoryEntries(t, parent)) {
		t.Fatal("symlink root rejection mutated filesystem state")
	}
}

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

// CLI-PATH-007: output targets must be regular files when they already exist.
func TestCLIPATH007RejectsNonRegularOutput(t *testing.T) {
	outputDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(outputDir, generatedFileName), 0o755); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	code := runGenerate([]string{"billing.gooo", "--out", outputDir}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitFailure || !strings.Contains(stderr.String(), "not a regular file") {
		t.Fatalf("non-regular output = code %d, stderr %q", code, stderr.String())
	}
}

// CLI-CAP-001: check refuses inputs above its bounded read size.
func TestCLICAP001RejectsOversizedCheckInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.gooo")
	data := []byte(strings.Repeat("x", int(maxInputBytes)+1))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	code := runCheck([]string{path}, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitFailure || !strings.Contains(stderr.String(), "maximum size") {
		t.Fatalf("oversized input = code %d, stderr %q", code, stderr.String())
	}
}

// CLI-CAP-002: diagnostics are bounded before they are emitted.
func TestCLICAP002RejectsExcessiveDiagnostics(t *testing.T) {
	diagnostics := make(syntax.Diagnostics, maxDiagnosticCount+1)
	parser := fixedDiagnosticsParser{diagnostics: diagnostics}
	var stderr bytes.Buffer
	code := runCheck([]string{"fixture.gooo"}, fixtureReader{source: validSource}, parser, &bytes.Buffer{}, &stderr)
	if code != exitFailure || stderr.String() != "gooo: diagnostic resource limit exceeded\n" {
		t.Fatalf("diagnostic cap = code %d, stderr %q", code, stderr.String())
	}
}

// CLI-CAP-003: output bytes and parser deadlines are bounded.
func TestCLICAP003BoundsOutputAndDeadline(t *testing.T) {
	t.Run("output", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), generatedFileName)
		err := writeGeneratedOutput(path, make([]byte, maxOutputBytes+1))
		if err == nil || !strings.Contains(err.Error(), "maximum size") {
			t.Fatalf("oversized output error = %v", err)
		}
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("oversized output created target: %v", err)
		}
	})
	t.Run("deadline", func(t *testing.T) {
		release := make(chan struct{})
		defer close(release)
		_, _, err := parseWithDeadline(blockingParser{release: release}, "fixture.gooo", validSource, 10*time.Millisecond)
		if !errors.Is(err, errCommandDeadline) {
			t.Fatalf("deadline error = %v", err)
		}
	})
}

// CLI-ATOMIC-001: successful output leaves no temporary sibling.
func TestCLIATOMIC001WritesAndCleansTemporaryOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, generatedFileName)
	if err := writeAtomicFile(path, []byte("atomic bytes\n")); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != generatedFileName {
		t.Fatalf("temporary output was not cleaned: %+v", entries)
	}
}

// CLI-ATOMIC-002: a failed rename preserves the old regular file and cleans up.
func TestCLIATOMIC002RenameFailurePreservesOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, generatedFileName)
	if err := os.WriteFile(path, []byte("old bytes\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	ops := defaultAtomicFileOps()
	ops.rename = func(string, string) error { return errors.New("simulated crash") }
	ops.syncDir = func(string) error { return nil }
	if err := writeAtomicFileWithOps(path, []byte("new bytes\n"), ops); err == nil {
		t.Fatal("rename failure was accepted")
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("old bytes\n")) || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) || len(entries) != 1 {
		t.Fatal("rename failure changed old output metadata or left a temporary file")
	}
}

type fixedDiagnosticsParser struct {
	diagnostics syntax.Diagnostics
}

func (p fixedDiagnosticsParser) ParseFile(string, string) (*syntax.File, syntax.Diagnostics) {
	return &syntax.File{}, p.diagnostics
}

type blockingParser struct {
	release <-chan struct{}
}

func (p blockingParser) ParseFile(string, string) (*syntax.File, syntax.Diagnostics) {
	<-p.release
	return &syntax.File{}, nil
}
