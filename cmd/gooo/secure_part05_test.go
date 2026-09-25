package main

import (
	"bytes"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"testing"
)

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
