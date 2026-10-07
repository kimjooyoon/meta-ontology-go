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
	if code := run(nil, &stdout, &stderr); code != exitUsage || stderr.String() != rootHelp {
		t.Fatalf("root usage = code %d, stderr %q", code, stderr.String())
	}
}

func TestHelpShowsQuickStartAndCommandGuide(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"--help"}, {"body-codegen", "--help"}, {"run", "--help"}} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != exitOK {
			t.Fatalf("help %v failed: code=%d stdout=%q stderr=%q", args, code, stdout.String(), stderr.String())
		}
		if stdout.Len() == 0 || stderr.Len() != 0 {
			t.Fatalf("help %v did not write a guide to stdout: stdout=%q stderr=%q", args, stdout.String(), stderr.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"help", "body-codegen"}, &stdout, &stderr); code != exitOK ||
		!strings.Contains(stdout.String(), "typed candidates declared by the plan") {
		t.Fatalf("body-codegen guide missing bounded model explanation: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"help", "run"}, &stdout, &stderr); code != exitOK ||
		!strings.Contains(stdout.String(), "reports its typed inputs and output") ||
		!strings.Contains(stdout.String(), "With an explicit input") ||
		!strings.Contains(stdout.String(), "value-plan path executes") ||
		!strings.Contains(stdout.String(), "value-plan path executes registered value operations") ||
		!strings.Contains(stdout.String(), "SOURCE_RUNTIME_BINDINGS_UNSUPPORTED") {
		t.Fatalf("run guide does not distinguish declaration and value execution: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestHelpRejectsUnknownTopicWithNextStep(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"help", "no-such-topic"}, &stdout, &stderr); code != exitUsage ||
		!strings.Contains(stderr.String(), "run `gooo help` for available topics") {
		t.Fatalf("unknown topic result: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
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

func TestRunInitRejectsInvalidLibraryModulePath(t *testing.T) {
	for _, module := range []string{"../escape", "Uppercase/Name", "example.org//empty"} {
		var stdout, stderr bytes.Buffer
		destination := filepath.Join(t.TempDir(), "library")
		code := run([]string{"init", "--template", "library", "--module", module, destination}, &stdout, &stderr)
		if code != exitUsage || !strings.Contains(stderr.String(), "invalid library module path") {
			t.Fatalf("invalid module path %q was accepted: code=%d stdout=%q stderr=%q", module, code, stdout.String(), stderr.String())
		}
		if _, err := os.Stat(destination); !os.IsNotExist(err) {
			t.Fatalf("invalid module path %q created a destination: stat err=%v", module, err)
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
