// Package semanticbinding contains the independent fixture oracle for the
// semantic-binding implementation. The fixtures remain literal and the
// execution cases call the implementation directly.
package semanticbinding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fixturePosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type fixtureLocation struct {
	Filename string          `json:"filename"`
	Start    fixturePosition `json:"start"`
	End      fixturePosition `json:"end"`
}

type fixtureRecord struct {
	Directive string           `json:"directive"`
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Subject   string           `json:"subject,omitempty"`
	Pressure  string           `json:"pressure,omitempty"`
	Location  *fixtureLocation `json:"location,omitempty"`
}

type fixtureExpectation struct {
	Status     string          `json:"status"`
	Canonical  string          `json:"canonical"`
	Diagnostic string          `json:"diagnostic,omitempty"`
	Records    []fixtureRecord `json:"records"`
}

var fixtureNames = []string{
	"valid_bind",
	"valid_obligation",
	"rename_before",
	"rename_after",
	"same_name_without_directive",
	"detached_comment",
	"unknown_field",
	"duplicate_field",
	"duplicate_id",
	"invalid_uri",
	"multi_name_var",
	"exact_spans",
	"canonical_permutation_a",
	"canonical_permutation_b",
	"filename_identity_a",
	"filename_identity_b",
}

func TestFixtureCorpusIsLiteralAndParseable(t *testing.T) {
	for _, name := range fixtureNames {
		t.Run(name, func(t *testing.T) {
			source, want := loadFixture(t, name)
			if _, diagnostics := parser.ParseFile(
				token.NewFileSet(), name+".go", source, parser.ParseComments,
			); diagnostics != nil {
				t.Fatalf("fixture is not valid Go: %v", diagnostics)
			}
			if want.Status != "accepted" && want.Status != "rejected" {
				t.Fatalf("status = %q, want accepted or rejected", want.Status)
			}
			if want.Records == nil {
				t.Fatal("expected records must be present as a literal array")
			}
			if want.Status == "rejected" && len(want.Records) != 0 {
				t.Fatalf("rejected fixture has records: %#v", want.Records)
			}
			for index, record := range want.Records {
				if record.Directive == "" || record.Name == "" || record.ID == "" {
					t.Fatalf("record[%d] is incomplete: %#v", index, record)
				}
			}
		})
	}
}

func TestTwoDeclarationsRetainLiteralExactSpans(t *testing.T) {
	source, want := loadFixture(t, "exact_spans")
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "exact_spans.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var declarations []*ast.GenDecl
	for _, declaration := range file.Decls {
		if group, ok := declaration.(*ast.GenDecl); ok && group.Tok == token.TYPE {
			declarations = append(declarations, group)
		}
	}
	if len(declarations) != len(want.Records) {
		t.Fatalf("type declarations = %d, want %d", len(declarations), len(want.Records))
	}
	for index, declaration := range declarations {
		location := want.Records[index].Location
		if location == nil {
			t.Fatalf("record[%d] has no literal location", index)
		}
		start := fileSet.PositionFor(declaration.Pos(), false)
		end := fileSet.PositionFor(declaration.End(), false)
		gotStart := fixturePosition{Offset: start.Offset, Line: start.Line, Column: start.Column}
		gotEnd := fixturePosition{Offset: end.Offset, Line: end.Line, Column: end.Column}
		if gotStart != location.Start || gotEnd != location.End || location.Filename != "exact_spans.go" {
			t.Fatalf("record[%d] location = %#v..%#v, want %#v..%#v", index, gotStart, gotEnd, location.Start, location.End)
		}
	}
}

func TestRenamePreservesLiteralIdentity(t *testing.T) {
	_, before := loadFixture(t, "rename_before")
	_, after := loadFixture(t, "rename_after")
	if len(before.Records) != 1 || len(after.Records) != 1 {
		t.Fatalf("rename records = %#v and %#v, want one record each", before.Records, after.Records)
	}
	if before.Records[0].Directive != after.Records[0].Directive || before.Records[0].ID != after.Records[0].ID {
		t.Fatalf("rename changed identity: before=%#v after=%#v", before.Records[0], after.Records[0])
	}
	if before.Records[0].Name == after.Records[0].Name {
		t.Fatalf("rename fixtures have the same display name: %q", before.Records[0].Name)
	}
}

func TestWhitespaceAndCommentPermutationHasLiteralCanonicalEquality(t *testing.T) {
	aSource, aWant := loadFixture(t, "canonical_permutation_a")
	bSource, bWant := loadFixture(t, "canonical_permutation_b")
	if bytes.Equal(aSource, bSource) {
		t.Fatal("permutation fixtures must differ in source presentation")
	}
	if aWant.Canonical == "" || aWant.Canonical != bWant.Canonical {
		t.Fatalf("canonical records differ: %q and %q", aWant.Canonical, bWant.Canonical)
	}
}

func TestFilenameChangeRetainsIdentityAndOnlyChangesLocation(t *testing.T) {
	aSource, aWant := loadFixture(t, "filename_identity_a")
	bSource, bWant := loadFixture(t, "filename_identity_b")
	if !bytes.Equal(aSource, bSource) {
		t.Fatal("filename fixtures must have identical source bytes")
	}
	if len(aWant.Records) != 1 || len(bWant.Records) != 1 {
		t.Fatalf("filename records = %#v and %#v, want one record each", aWant.Records, bWant.Records)
	}
	aRecord, bRecord := aWant.Records[0], bWant.Records[0]
	if aRecord.Directive != bRecord.Directive || aRecord.Name != bRecord.Name ||
		aRecord.ID != bRecord.ID || aRecord.Subject != bRecord.Subject ||
		aRecord.Pressure != bRecord.Pressure {
		t.Fatalf("filename changed semantic record: a=%#v b=%#v", aRecord, bRecord)
	}
	if aRecord.Location == nil || bRecord.Location == nil || aRecord.Location.Filename == bRecord.Location.Filename {
		t.Fatalf("filename locations were not retained independently: a=%#v b=%#v", aRecord.Location, bRecord.Location)
	}
	if aRecord.Location.Start != bRecord.Location.Start || aRecord.Location.End != bRecord.Location.End {
		t.Fatalf("filename changed non-location span data: a=%#v b=%#v", aRecord.Location, bRecord.Location)
	}
}

func TestSemanticBindingImplementationContract(t *testing.T) {
	t.Run("valid bind", func(t *testing.T) { assertFixture(t, "valid_bind") })
	t.Run("valid obligation", func(t *testing.T) { assertFixture(t, "valid_obligation") })
	t.Run("rename preserving ID", func(t *testing.T) {
		before := assertFixture(t, "rename_before")
		after := assertFixture(t, "rename_after")
		if len(before.records) != 1 || len(after.records) != 1 {
			t.Fatalf("rename records = %#v and %#v, want one record each", before.records, after.records)
		}
		if before.records[0].ID != after.records[0].ID ||
			before.records[0].Directive != after.records[0].Directive {
			t.Fatalf("rename changed semantic identity: before=%#v after=%#v", before.records, after.records)
		}
		if before.records[0].Name == after.records[0].Name {
			t.Fatalf("rename fixtures have the same display name: %q", before.records[0].Name)
		}
		if before.result.Bindings[0].DeclarationKey == after.result.Bindings[0].DeclarationKey {
			t.Fatalf("rename did not change the Go declaration key: %q", before.result.Bindings[0].DeclarationKey)
		}
		if before.result.Canonical() != after.result.Canonical() ||
			before.result.StableHash() != after.result.StableHash() {
			t.Fatalf("rename changed canonical semantic output: before=%s after=%s", before.result.StableHash(), after.result.StableHash())
		}
	})
	t.Run("same name without directive", func(t *testing.T) { assertFixture(t, "same_name_without_directive") })
	t.Run("detached comment", func(t *testing.T) { assertFixture(t, "detached_comment") })
	t.Run("unknown field", func(t *testing.T) { assertFixture(t, "unknown_field") })
	t.Run("duplicate field", func(t *testing.T) { assertFixture(t, "duplicate_field") })
	t.Run("duplicate ID", func(t *testing.T) { assertFixture(t, "duplicate_id") })
	t.Run("invalid URI", func(t *testing.T) { assertFixture(t, "invalid_uri") })
	t.Run("multi-name var declaration", func(t *testing.T) { assertFixture(t, "multi_name_var") })
	t.Run("two declarations with exact spans", func(t *testing.T) { assertFixture(t, "exact_spans") })
	t.Run("whitespace/comment canonical equality", func(t *testing.T) {
		a := assertFixture(t, "canonical_permutation_a")
		b := assertFixture(t, "canonical_permutation_b")
		if a.oracleCanonical != b.oracleCanonical || a.result.Canonical() != b.result.Canonical() ||
			a.result.StableHash() != b.result.StableHash() {
			t.Fatalf("presentation permutation changed canonical output: a=%q/%s b=%q/%s",
				a.oracleCanonical, a.result.StableHash(), b.oracleCanonical, b.result.StableHash())
		}
	})
	t.Run("filename identity versus location", func(t *testing.T) {
		a := assertFixture(t, "filename_identity_a")
		b := assertFixture(t, "filename_identity_b")
		if a.oracleCanonical != b.oracleCanonical || len(a.records) != 1 || len(b.records) != 1 {
			t.Fatalf("filename fixtures changed semantic records: a=%#v b=%#v", a.records, b.records)
		}
		if a.records[0].ID != b.records[0].ID || a.records[0].Name != b.records[0].Name {
			t.Fatalf("filename fixtures changed identity: a=%#v b=%#v", a.records, b.records)
		}
		if a.records[0].Location == nil || b.records[0].Location == nil ||
			a.records[0].Location.Filename == b.records[0].Location.Filename {
			t.Fatalf("filename locations were not retained independently: a=%#v b=%#v", a.records, b.records)
		}
		if a.records[0].Location.Start != b.records[0].Location.Start ||
			a.records[0].Location.End != b.records[0].Location.End {
			t.Fatalf("filename changed non-location span data: a=%#v b=%#v", a.records, b.records)
		}
		if a.result.Canonical() != b.result.Canonical() || a.result.StableHash() != b.result.StableHash() {
			t.Fatalf("filename relocation changed canonical semantic output: a=%s b=%s", a.result.StableHash(), b.result.StableHash())
		}
	})
}

func TestNonExactDirectiveMarkerDoesNotBind(t *testing.T) {
	source := []byte("package billing\n\n// gooo:bind id=\"billing://entity/order\" role=\"HANDWRITTEN_IMPL\"\ntype Order struct{}\n")
	result, err := Extract(Input{Sources: []SourceFile{{
		Filename: "non_exact_marker.go", PackagePath: "billing", Source: source,
	}}})
	if err != nil {
		t.Fatalf("Extract returned error for an ordinary comment: %v", err)
	}
	if result.Status != StatusBound || len(result.Bindings) != 0 || len(result.Obligations) != 0 {
		t.Fatalf("result = %#v, want no semantic records", result)
	}
}

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
		if err != nil {
			t.Fatalf("Extract returned error: %v", err)
		}
		if result.Status != StatusBound || result.FullSuiteFallback || len(result.Unknowns) != 0 {
			t.Fatalf("result = %#v, want complete BOUND result", result)
		}
	} else {
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
		if typed.Code != oracleCode(want.Diagnostic) {
			t.Fatalf("diagnostic code = %q, want oracle %q", typed.Code, want.Diagnostic)
		}
		if result.Unknowns[0].Code != typed.Code || result.Unknowns[0].Span != typed.Span ||
			!result.Unknowns[0].FullSuiteFallback {
			t.Fatalf("UNKNOWN evidence = %#v, want the source-backed error", result.Unknowns[0])
		}
		if len(result.Bindings) != 0 || len(result.Obligations) != 0 {
			t.Fatalf("rejected result retained partial records: %#v", result)
		}
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
	records := make([]fixtureRecord, 0, len(result.Bindings)+len(result.Obligations))
	for _, binding := range result.Bindings {
		records = append(records, fixtureRecord{
			Directive: "bind", Name: nameAt(binding.Span), ID: binding.ID,
			Location: locationForSpan(binding.Span),
		})
	}
	for _, obligation := range result.Obligations {
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

func canonicalFixture(records []fixtureRecord) string {
	parts := make([]string, 0, len(records))
	for _, record := range records {
		if record.Directive == "obligation" {
			parts = append(parts, fmt.Sprintf("obligation|%s|%s|%s", record.ID, record.Subject, record.Pressure))
			continue
		}
		parts = append(parts, fmt.Sprintf("bind|%s", record.ID))
	}
	return strings.Join(parts, "\n")
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

func loadFixture(t *testing.T, name string) ([]byte, fixtureExpectation) {
	t.Helper()
	root := "testdata"
	source, err := os.ReadFile(filepath.Join(root, name+".go"))
	if err != nil {
		t.Fatal(err)
	}
	wantBytes, err := os.ReadFile(filepath.Join(root, name+".want.json"))
	if err != nil {
		t.Fatal(err)
	}
	var want fixtureExpectation
	if err := json.Unmarshal(wantBytes, &want); err != nil {
		t.Fatalf("decode %s.want.json: %v", name, err)
	}
	return source, want
}
