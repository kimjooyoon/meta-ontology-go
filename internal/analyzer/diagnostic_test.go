package analyzer

import (
	"reflect"
	"testing"
)

func TestInvalidSemanticAnnotationsProduceStableDiagnostics(t *testing.T) {
	files := []SourceFile{
		{
			Filename: "z-invalid.go",
			Source: []byte(`package billing

//gooo:semantic kind=agent id="billing://entity/agent"
type Agent struct{}
`),
		},
		{
			Filename: "a-invalid.go",
			Source: []byte(`package billing

//gooo:semantic entity
type MissingID struct{}

//gooo:semantic entity id="billing://entity/one" kind=activity
type ConflictingKind struct{}
`),
		},
	}

	first, err := AnalyzePackage(files, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := AnalyzePackage([]SourceFile{files[1], files[0]}, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Diagnostics, reversed.Diagnostics) {
		t.Fatalf("diagnostics depend on source order:\nfirst=%#v\nreversed=%#v", first.Diagnostics, reversed.Diagnostics)
	}
	if len(first.Diagnostics) != 3 {
		t.Fatalf("diagnostics = %#v, want three invalid annotations", first.Diagnostics)
	}
	if got := first.Diagnostics[0].Code; got != DiagInvalidAnnotation {
		t.Fatalf("diagnostic[0] code = %q, want %q", got, DiagInvalidAnnotation)
	}
	if got := first.Diagnostics[1].Code; got != DiagConflictingAnnotation {
		t.Fatalf("diagnostic[1] code = %q, want %q", got, DiagConflictingAnnotation)
	}
	if first.Diagnostics[0].Span.Filename != "a-invalid.go" || first.Diagnostics[2].Span.Filename != "z-invalid.go" {
		t.Fatalf("diagnostics are not filename ordered: %#v", first.Diagnostics)
	}
	if len(first.Registrations) != 0 || len(first.Delta.Added) != 0 {
		t.Fatalf("invalid annotations crossed semantic boundary: registrations=%#v delta=%#v", first.Registrations, first.Delta)
	}
}

func TestValidSemanticAnnotationDoesNotEmitDiagnostic(t *testing.T) {
	result, err := AnalyzeSource("valid.go", []byte(`package billing

//gooo:semantic entity id="billing://entity/order"
type Order struct{}
`), NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if result.Diagnostics.HasErrors() {
		t.Fatalf("valid annotation diagnostics = %#v", result.Diagnostics)
	}
	if len(result.Registrations) != 1 {
		t.Fatalf("registrations = %#v, want one", result.Registrations)
	}
}
