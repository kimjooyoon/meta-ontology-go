package semanticbinding

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

type fixtureObservation struct {
	result          Result
	records         []fixtureRecord
	oracleCanonical string
}

func assertFixture(t *testing.T, name string) fixtureObservation {
	t.Helper()
	source, want := loadFixture(t, name)
	result, err := Extract(Input{Sources: []SourceFile{{
		Filename: name + ".go", PackagePath: "billing", Source: source,
	}}})
	if want.Status == "accepted" {
		assertAcceptedFixture(t, result, err)
	} else {
		assertRejectedFixture(t, result, err, want.Diagnostic)
		return fixtureObservation{result: result}
	}
	records := recordsForResult(t, name, source, result)
	if len(records) != len(want.Records) {
		t.Fatalf("records = %#v, want oracle records %#v", records, want.Records)
	}
	for index := range want.Records {
		assertRecord(t, index, records[index], want.Records[index])
	}
	oracleCanonical := canonicalFixture(want.Records)
	if oracleCanonical != want.Canonical {
		t.Fatalf("oracle canonical = %q, want literal %q", oracleCanonical, want.Canonical)
	}
	if result.Digest == "" || result.Digest != result.CanonicalDigest || result.Digest != result.StableHash() {
		t.Fatalf("result digest fields are inconsistent: %#v", result)
	}
	return fixtureObservation{result: result, records: records, oracleCanonical: oracleCanonical}
}

func assertAcceptedFixture(t *testing.T, result Result, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if result.Status != StatusBound || result.FullSuiteFallback || len(result.Unknowns) != 0 {
		t.Fatalf("result = %#v, want complete BOUND result", result)
	}
}

func assertRejectedFixture(t *testing.T, result Result, err error, diagnostic string) {
	t.Helper()
	if err == nil {
		t.Fatal("Extract accepted a rejected oracle fixture")
	}
	if result.Status != StatusUnknown || !result.FullSuiteFallback || len(result.Unknowns) != 1 {
		t.Fatalf("result = %#v, want UNKNOWN with full-suite fallback", result)
	}
	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatalf("Extract error = %T %v, want *Error", err, err)
	}
	if typed.Code != oracleCode(diagnostic) {
		t.Fatalf("diagnostic code = %q, want oracle %q", typed.Code, diagnostic)
	}
	if result.Unknowns[0].Code != typed.Code || result.Unknowns[0].Span != typed.Span ||
		!result.Unknowns[0].FullSuiteFallback {
		t.Fatalf("UNKNOWN evidence = %#v, want the source-backed error", result.Unknowns[0])
	}
	if len(result.Bindings) != 0 || len(result.Obligations) != 0 {
		t.Fatalf("rejected result retained partial records: %#v", result)
	}
}

func recordsForResult(t *testing.T, name string, source []byte, result Result) []fixtureRecord {
	t.Helper()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, name+".go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse fixture for oracle comparison: %v", err)
	}
	nameAt := func(span Span) string {
		for _, declaration := range file.Decls {
			if spanFor(fileSet, declaration) == span {
				return declarationName(declaration)
			}
			if group, ok := declaration.(*ast.GenDecl); ok {
				for _, specification := range group.Specs {
					if spanFor(fileSet, specification) == span {
						return declarationName(specification)
					}
				}
			}
		}
		t.Fatalf("no declaration for result span %s", span)
		return ""
	}
	return appendBindingRecords(result.Bindings, result.Obligations, nameAt)
}

func appendBindingRecords(bindings []Binding, obligations []Obligation, nameAt func(Span) string) []fixtureRecord {
	records := make([]fixtureRecord, 0, len(bindings)+len(obligations))
	for _, binding := range bindings {
		records = append(records, fixtureRecord{
			Directive: "bind", Name: nameAt(binding.Span), ID: binding.ID,
			Location: locationForSpan(binding.Span),
		})
	}
	for _, obligation := range obligations {
		records = append(records, fixtureRecord{
			Directive: "obligation", Name: nameAt(obligation.Span), ID: obligation.ID,
			Subject: obligation.Subject, Pressure: obligation.Pressure,
			Location: locationForSpan(obligation.Span),
		})
	}
	return records
}

func declarationName(node ast.Node) string {
	switch current := node.(type) {
	case *ast.FuncDecl:
		return current.Name.Name
	case *ast.GenDecl:
		if len(current.Specs) == 1 {
			return declarationName(current.Specs[0])
		}
	case *ast.TypeSpec:
		return current.Name.Name
	case *ast.ValueSpec:
		if len(current.Names) == 1 {
			return current.Names[0].Name
		}
	}
	return ""
}

func locationForSpan(span Span) *fixtureLocation {
	return &fixtureLocation{
		Filename: span.Filename,
		Start:    fixturePosition{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
		End:      fixturePosition{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
	}
}

func assertRecord(t *testing.T, index int, got, want fixtureRecord) {
	t.Helper()
	if got.Directive != want.Directive || got.Name != want.Name || got.ID != want.ID ||
		got.Subject != want.Subject || got.Pressure != want.Pressure {
		t.Fatalf("record[%d] = %#v, want oracle %#v", index, got, want)
	}
	if want.Location != nil && (got.Location == nil || *got.Location != *want.Location) {
		t.Fatalf("record[%d] location = %#v, want oracle %#v", index, got.Location, want.Location)
	}
}

func oracleCode(diagnostic string) Code {
	switch diagnostic {
	case "detached-comment":
		return CodeDetachedComment
	case "unknown-field":
		return CodeUnknownField
	case "duplicate-field":
		return CodeDuplicateField
	case "duplicate-id":
		return CodeDuplicateID
	case "invalid-uri":
		return CodeInvalidIdentity
	case "multi-name-declaration":
		return CodeAmbiguousBinding
	default:
		return Code(diagnostic)
	}
}
