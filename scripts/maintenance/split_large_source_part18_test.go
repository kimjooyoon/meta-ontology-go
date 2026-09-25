package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorPreservesBuildTestAndAliasIdentity(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "sample_linux_test.go")
	source := `//go:build linux

// Package sample exercises source identity.
package sample

import alias "strings"

func Alpha() string {
	return alias.TrimSpace(" a ")
}

func Beta() string {
	return alias.TrimSpace(" b ")
}
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	rewritten, err := refactor(sourcePath, options{maxLines: 12, maxEntries: 10, write: true})
	if err != nil {
		t.Fatal(err)
	}
	if rewritten != 1 {
		t.Fatalf("rewritten = %d, want 1", rewritten)
	}
	for part := 1; part <= 2; part++ {
		path, err := generatedPath(sourcePath, part)
		if err != nil {
			t.Fatal(err)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(contents)
		if !strings.Contains(text, "//go:build linux") || !strings.Contains(text, `alias "strings"`) {
			t.Fatalf("generated identity missing from %s:\n%s", path, text)
		}
	}
	if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
		t.Fatalf("source should be replaced, stat error = %v", err)
	}
}
