package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunAnalyzeStableAcrossSingleGeneratedGoOutputRoots(t *testing.T) {
	authority, generated := billingAnalyzeFiles(t, billingAnalyzeAuthority)
	generatedSource, err := os.ReadFile(generated)
	if err != nil {
		t.Fatalf("read generated source: %v", err)
	}

	firstPath := filepath.Join(t.TempDir(), "semantic.gooo.go")
	secondPath := filepath.Join(t.TempDir(), "semantic.gooo.go")
	if err := os.WriteFile(firstPath, generatedSource, 0o600); err != nil {
		t.Fatalf("write first generated source: %v", err)
	}
	if err := os.WriteFile(secondPath, generatedSource, 0o600); err != nil {
		t.Fatalf("write second generated source: %v", err)
	}

	first, firstCode, firstErr := runAnalyzePaths(authority, firstPath)
	if firstErr != "" || firstCode != exitOK {
		t.Fatalf("first analysis failed: code=%d err=%v", firstCode, firstErr)
	}
	second, secondCode, secondErr := runAnalyzePaths(authority, secondPath)
	if secondErr != "" || secondCode != exitOK {
		t.Fatalf("second analysis failed: code=%d err=%v", secondCode, secondErr)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("single-file analysis changed with the generated source root")
	}
}
