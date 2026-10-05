package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDispatchesCheckAndUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check"}, &stdout, &stderr); code != exitUsage || !strings.Contains(stderr.String(), "usage: gooo check [--semantic]") {
		t.Fatalf("check usage = code %d, stderr %q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(nil, &stdout, &stderr); code != exitUsage || stderr.String() != "usage: gooo <init|run|compare|propose-repair|consume-repair|revise-source|revise-from-handoff|evaluate-revision|verify-revision-contract|run-accepted-revision|compare-accepted-revision|stage-accepted-revision|profile|debug|test|emit|receipt-schema|check|decide|generate|body-codegen|body-path-stream|body-path-run|body-execute|body-realize|body-compose|completeness-delta|body-context|roundtrip|query|inspect|graph|claim|analyze|format|fix|provenance|selective-ci|invoke|lsp|version> [args]\n" {
		t.Fatalf("root usage = code %d, stderr %q", code, stderr.String())
	}
}

func TestRunInitCreatesCheckableModelReadyProject(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "starter")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", directory}, &stdout, &stderr); code != exitOK {
		t.Fatalf("init failed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	for name := range starterFiles {
		if _, err := os.Stat(filepath.Join(directory, name)); err != nil {
			t.Fatalf("init did not create %s: %v", name, err)
		}
	}
	var checked, diagnostics bytes.Buffer
	if code := run([]string{"check", filepath.Join(directory, "main.gooo")}, &checked, &diagnostics); code != exitOK {
		t.Fatalf("generated source does not pass gooo check: code=%d stdout=%q stderr=%q", code, checked.String(), diagnostics.String())
	}
	if !strings.Contains(stdout.String(), "gooo check main.gooo") {
		t.Fatalf("init output lacks the next action: %q", stdout.String())
	}
}

func TestRunInitRejectsExistingDestinationWithoutChangingIt(t *testing.T) {
	directory := t.TempDir()
	marker := filepath.Join(directory, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", directory}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("init accepted an existing directory: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "keep" {
		t.Fatalf("existing destination changed: content=%q err=%v", got, err)
	}
}

func TestRunInitRequiresOneDestination(t *testing.T) {
	for _, args := range [][]string{{"init"}, {"init", "first", "second"}, {"init", "-"}} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != exitUsage || stderr.String() != initUsage+"\n" {
			t.Fatalf("unexpected usage result for %v: code=%d stdout=%q stderr=%q", args, code, stdout.String(), stderr.String())
		}
	}
}

type fixtureReader struct {
	source string
	err    error
}

func (r fixtureReader) ReadFile(string) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return []byte(r.source), nil
}

type recordingParser struct {
	filename string
	source   string
}

func (p *recordingParser) ParseFile(filename, source string) (*syntax.File, syntax.Diagnostics) {
	p.filename = filename
	p.source = source
	return syntax.ParseFile(filename, source)
}

const validSource = `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Order
`
