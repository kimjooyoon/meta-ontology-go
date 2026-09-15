package main

import (
	"bytes"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
