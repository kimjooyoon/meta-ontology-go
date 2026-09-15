package main

import (
	"bytes"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
	"time"
)

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
