// Package semanticbinding contains the independent fixture contract for the
// semantic-binding implementation. The implementation is absent at the
// authority commit, so execution cases remain explicit dependency-local
// NOT_RUN instead of using a test-only substitute.
package semanticbinding

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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
	Field     string           `json:"field,omitempty"`
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
		aRecord.ID != bRecord.ID || aRecord.Field != bRecord.Field {
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
	for _, name := range []string{
		"valid bind",
		"valid obligation",
		"rename preserving ID",
		"same name without directive",
		"detached comment",
		"unknown field",
		"duplicate field",
		"duplicate ID",
		"invalid URI",
		"multi-name var declaration",
		"two declarations with exact spans",
		"whitespace/comment canonical equality",
		"filename identity versus location",
	} {
		t.Run(name, func(t *testing.T) {
			t.Skip("dependency-local NOT_RUN: semanticbinding implementation is absent at " +
				"authority 9375ca0649a78feafd69f3ae22dd08d976add7c0")
		})
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
