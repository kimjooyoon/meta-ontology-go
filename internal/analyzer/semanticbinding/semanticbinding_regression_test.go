package semanticbinding

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

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
		aRecord.ID != bRecord.ID || aRecord.Subject != bRecord.Subject || aRecord.Pressure != bRecord.Pressure {
		t.Fatalf("filename changed semantic record: a=%#v b=%#v", aRecord, bRecord)
	}
	if aRecord.Location == nil || bRecord.Location == nil || aRecord.Location.Filename == bRecord.Location.Filename {
		t.Fatalf("filename locations were not retained independently: a=%#v b=%#v", aRecord.Location, bRecord.Location)
	}
	if aRecord.Location.Start != bRecord.Location.Start || aRecord.Location.End != bRecord.Location.End {
		t.Fatalf("filename changed non-location span data: a=%#v b=%#v", aRecord.Location, bRecord.Location)
	}
}
