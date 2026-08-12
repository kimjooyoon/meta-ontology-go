package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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
	code := runCheck([]string{"--semantic", filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
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

func TestSemanticCheckPropagatesInvalidRelationDeterministically(t *testing.T) {
	lower := func(*syntax.File) (semantic.IR, error) { return semantic.IR{}, semantic.ErrUnknownRelation }
	first, second := bytes.Buffer{}, bytes.Buffer{}
	_, firstErr := semanticCheckIRWithLowerer(&syntax.File{}, time.Second, lower)
	_, secondErr := semanticCheckIRWithLowerer(&syntax.File{}, time.Second, lower)
	if !errors.Is(firstErr, semantic.ErrUnknownRelation) || !errors.Is(secondErr, semantic.ErrUnknownRelation) {
		t.Fatalf("relation errors = %v/%v", firstErr, secondErr)
	}
	if !reportSemanticDiagnostic("fixture.gooo", &syntax.File{}, firstErr, &first) ||
		!reportSemanticDiagnostic("fixture.gooo", &syntax.File{}, secondErr, &second) || first.String() != second.String() {
		t.Fatalf("relation diagnostics are not deterministic: %q/%q", first.String(), second.String())
	}
	if !bytes.Contains(first.Bytes(), []byte("semantic.invalid-relation")) {
		t.Fatalf("relation diagnostic = %q", first.String())
	}
}

func TestSemanticCheckUsesVersionedReadOnlyBoundary(t *testing.T) {
	if semanticCheckSchemaVersion != "gooo-semantic-check/v1" {
		t.Fatalf("semantic check schema = %q", semanticCheckSchemaVersion)
	}
	directory := t.TempDir()
	filename := filepath.Join(directory, "valid.gooo")
	if err := os.WriteFile(filename, []byte(validSource), 0o640); err != nil {
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
	code := runCheck([]string{"--semantic", filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stdout.String() != "ok: "+filename+"\n" || stderr.String() != deferredCheckProvenance+"\n" {
		t.Fatalf("semantic check = code %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
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
		t.Fatal("semantic check changed valid input")
	}
	if afterEntries := directoryEntries(t, directory); !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("semantic check changed directory entries: before=%v after=%v", beforeEntries, afterEntries)
	}
}

func TestSemanticCheckDeadlineBoundsLowerer(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := semanticCheckIRWithLowerer(&syntax.File{}, 10*time.Millisecond, func(*syntax.File) (semantic.IR, error) {
			close(started)
			<-release
			return semantic.IR{}, nil
		})
		done <- err
	}()
	<-started
	select {
	case err := <-done:
		if !errors.Is(err, errCommandDeadline) {
			t.Fatalf("deadline error = %v, want %v", err, errCommandDeadline)
		}
	case <-time.After(time.Second):
		t.Fatal("semantic check exceeded bounded deadline")
	}
	close(release)
}

func TestSemanticDiagnosticClassifiesDeadline(t *testing.T) {
	if code := semanticDiagnosticCode(errCommandDeadline); code != "semantic.deadline" {
		t.Fatalf("deadline diagnostic code = %q", code)
	}
}

func TestSemanticValidationDiagnosticsAreCanonical(t *testing.T) {
	span := syntax.Span{
		Filename: "fixture.gooo",
		Start:    syntax.Position{Offset: 0, Line: 1, Column: 1},
		End:      syntax.Position{Offset: 4, Line: 1, Column: 5},
	}
	issues := []semantic.ValidationIssue{
		{Code: "relation-kind", Message: "kind mismatch", Subject: "billing://activity/pay", Object: "billing://entity/order"},
		{Code: "missing-object", Message: "object is not declared", Subject: "billing://activity/pay", Object: "billing://entity/missing"},
	}
	first, err := formatSemanticDiagnostics(span, &semantic.ValidationErrors{Issues: issues})
	if err != nil {
		t.Fatal(err)
	}
	issues[0], issues[1] = issues[1], issues[0]
	second, err := formatSemanticDiagnostics(span, &semantic.ValidationErrors{Issues: issues})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("semantic validation diagnostics changed with insertion order:\nfirst=%ssecond=%s", first, second)
	}
	want := "fixture.gooo:1:1-1:5: error semantic.invalid-endpoint: object is not declared\n" +
		"fixture.gooo:1:1-1:5: error semantic.invalid-kind: kind mismatch\n"
	if string(first) != want {
		t.Fatalf("semantic validation diagnostics = %q, want %q", first, want)
	}
}

func TestInspectAndAnalyzeValidateLowererResults(t *testing.T) {
	invalidLowerer := func(*syntax.File) (semantic.IR, error) { return semantic.IR{}, nil }
	var inspectOut, inspectErr bytes.Buffer
	inspectCode := runInspectWithLowerer([]string{"fixture.gooo"}, fixtureReader{source: sourceOrderA}, SyntaxSourceParser{}, &inspectOut, &inspectErr, invalidLowerer)
	if inspectCode != exitFailure || inspectOut.Len() != 0 || !bytes.Contains(inspectErr.Bytes(), []byte("semantic.invalid")) {
		t.Fatalf("inspect accepted invalid IR = code %d, stdout=%q, stderr=%q", inspectCode, inspectOut.String(), inspectErr.String())
	}

	var analyzeOut, analyzeErr bytes.Buffer
	analyzeCode := runAnalyzeWithLowerer([]string{"fixture.gooo"}, fixtureReader{source: sourceOrderA}, SyntaxSourceParser{}, &analyzeOut, &analyzeErr, invalidLowerer)
	if analyzeCode != exitFailure || analyzeErr.Len() != 0 {
		t.Fatalf("analyze accepted invalid IR = code %d, stderr=%q", analyzeCode, analyzeErr.String())
	}
	plan := decodeFixPlan(t, analyzeOut.Bytes())
	if plan.Status != fixPlanSemanticInvalid || len(plan.Diagnostics) != 1 || plan.Diagnostics[0].Code != "semantic.lowering" {
		t.Fatalf("invalid IR analyze plan = %#v", plan)
	}
}

func checkFixture(t *testing.T, source string) (stdout, stderr string, code int) {
	t.Helper()
	var out, err bytes.Buffer
	code = runCheck([]string{"--semantic", "billing.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &out, &err)
	return out.String(), err.String(), code
}
