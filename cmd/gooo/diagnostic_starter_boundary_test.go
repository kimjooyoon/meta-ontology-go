package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticStarterBindsModuleAndPreservesExistingProject(t *testing.T) {
	root := filepath.Join(t.TempDir(), "diagnostic")
	var stdout, stderr bytes.Buffer
	args := []string{"init", "--template", "diagnostic", "--module", "example.org/acme/triage", root}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 6 {
		t.Fatal("expected six self-contained starter files", entries, err)
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil || strings.Contains(string(data), "{{") || strings.Contains(string(data), ".fixture") {
			t.Fatal("unexpanded starter file", entry.Name(), err, string(data))
		}
	}
	source := filepath.Join(root, "diagnostics.gooo")
	data, err := os.ReadFile(source)
	if err != nil || !strings.Contains(string(data), "gooo://example.org/acme/triage/diagnostics/diagnostic") {
		t.Fatal("module did not bind semantic identity", err, string(data))
	}
	if code := run(args, &stdout, &stderr); code != exitFailure {
		t.Fatal("existing diagnostic project was accepted", code)
	}
	after, err := os.ReadFile(source)
	if err != nil || !bytes.Equal(data, after) {
		t.Fatal("existing source changed", err)
	}
}

func TestDiagnosticStarterRejectsInvalidModuleBeforeWriting(t *testing.T) {
	for _, module := range []string{"../escape", "Upper/Name", "example.org//empty"} {
		root := filepath.Join(t.TempDir(), "diagnostic")
		var stdout, stderr bytes.Buffer
		code := run([]string{"init", "--template", "diagnostic", "--module", module, root}, &stdout, &stderr)
		if code != exitUsage || !strings.Contains(stderr.String(), "invalid diagnostic module path") {
			t.Fatal(code, stderr.String())
		}
		if _, err := os.Stat(root); !os.IsNotExist(err) {
			t.Fatal("invalid module created a project", err)
		}
	}
}
