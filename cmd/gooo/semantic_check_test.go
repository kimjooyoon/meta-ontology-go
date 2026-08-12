package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestRunCheckRejectsUnknownEndpointDeterministically(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Missing) -> Order
`
	firstOut, firstErr, firstCode := checkFixture(t, source)
	secondOut, secondErr, secondCode := checkFixture(t, source)
	if firstCode != exitFailure || secondCode != exitFailure || firstOut != "" || secondOut != "" {
		t.Fatalf("endpoint check = %d/%d, stdout=%q/%q", firstCode, secondCode, firstOut, secondOut)
	}
	if firstErr != secondErr || !bytes.Contains([]byte(firstErr), []byte("semantic.invalid-endpoint")) {
		t.Fatalf("endpoint diagnostics are not deterministic or classified: first=%q second=%q", firstErr, secondErr)
	}
}

func TestRunCheckRejectsInvalidKindWithoutMutation(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(PayOrder) -> Order
`
	directory := t.TempDir()
	filename := filepath.Join(directory, "invalid.gooo")
	if err := os.WriteFile(filename, []byte(source), 0o640); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeEntries := directoryEntries(t, directory)
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stdout.Len() != 0 || !bytes.Contains(stderr.Bytes(), []byte("semantic.invalid-kind")) {
		t.Fatalf("kind check = code %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	after, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || !os.SameFile(beforeInfo, afterInfo) || beforeInfo.Mode() != afterInfo.Mode() || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("semantic check changed invalid input")
	}
	if afterEntries := directoryEntries(t, directory); !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("semantic check changed directory entries: before=%v after=%v", beforeEntries, afterEntries)
	}
}

func TestRunCheckSeparatesSemanticValidationFromProvenance(t *testing.T) {
	stdout, stderr, code := checkFixture(t, validSource)
	if code != exitOK || stdout != "ok: billing.gooo\n" || stderr != deferredCheckProvenance+"\n" {
		t.Fatalf("check acceptance = code %d, stdout=%q, stderr=%q", code, stdout, stderr)
	}
}

func TestSemanticDiagnosticClassifiesInvalidRelation(t *testing.T) {
	if code := semanticDiagnosticCode(semantic.ErrUnknownRelation); code != "semantic.invalid-relation" {
		t.Fatalf("invalid relation code = %q", code)
	}
}

func checkFixture(t *testing.T, source string) (stdout, stderr string, code int) {
	t.Helper()
	var out, err bytes.Buffer
	code = runCheck([]string{"billing.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &out, &err)
	return out.String(), err.String(), code
}
