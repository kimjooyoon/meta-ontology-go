package bidir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRecordReferenceRenamePreservesIdentityAndSourceForm(t *testing.T) {
	for _, spelling := range []string{"Line", `"totals:Line"`, `"urn:record-items:line"`} {
		for _, reuseName := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/reuse=%t", spelling, reuseName), func(t *testing.T) {
				checkRecordReferenceRename(t, spelling, reuseName)
			})
		}
	}
}

func checkRecordReferenceRename(t *testing.T, spelling string, reuseName bool) {
	t.Helper()
	document, model, support := recordRenameFixture(t, spelling)
	original, snapshot := model.Clone(), recordRenameSnapshot(t, document)
	renameFixtureRecords(&model, reuseName)
	input := model.Clone()
	written, err := PutWithEntityFieldsSupport(document, model, support)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input, model) || !bytes.Equal(snapshot, recordRenameSnapshot(t, document)) {
		t.Fatal("Put mutated the caller model or original source document")
	}
	assertRecordRenameDeclarations(t, document, original, written, spelling)
	observed, err := GetWithEntityFieldsSupport(written, support)
	if err != nil || !SemanticEquivalent(original, observed) {
		t.Fatal("rename changed semantic identity", err)
	}
	if err := CheckPutGetWithEntityFieldsSupport(document, model, support); err != nil {
		t.Fatal(err)
	}
	if err := CheckGetPutWithEntityFieldsSupport(written, support); err != nil {
		t.Fatal(err)
	}
	assertRenamedRecordSourceReload(t, written, support)
}

func recordRenameSnapshot(t *testing.T, document Document) []byte {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func renameFixtureRecords(model *Model, reuseName bool) {
	for i := range model.Nodes {
		if model.Nodes[i].ID == "urn:record-items:line" {
			model.Nodes[i].Name = "Product"
		}
		if reuseName && model.Nodes[i].ID == "urn:record-items:other" {
			model.Nodes[i].Name = "Line"
		}
	}
}

func assertRecordRenameDeclarations(t *testing.T, source Document, original Model, written Document, spelling string) {
	t.Helper()
	want := strings.Trim(strings.Replace(spelling, "Line", "Product", 1), `"`)
	if strings.Contains(spelling, "urn:") {
		want = "urn:record-items:line"
	}
	nodes := nodeMap(original.Nodes)
	for i, declaration := range written.Declarations {
		before := source.Declarations[i]
		if declaration.ID != before.ID || declaration.Span != before.Span {
			t.Fatal("declaration identity, order or origin changed")
		}
		for j, field := range declaration.Fields {
			previous := nodes[declaration.ID].Fields[j]
			if declaration.ID == "urn:record-items:basket" {
				assertRenamedRecordField(t, field, previous, want)
			} else if !reflect.DeepEqual(field, previous) {
				t.Fatal("unrelated field changed", field)
			}
		}
	}
}

func assertRenamedRecordField(t *testing.T, field, previous Field, want string) {
	t.Helper()
	if field.TypeRefUse.Spelling != want || field.TypeRefUse.ResolvedID != "urn:record-items:line" {
		t.Fatal("renamed reference lost its source form or stable target", field)
	}
	comparison := field.clone()
	comparison.TypeRef, comparison.TypeRefUse = previous.TypeRef, previous.TypeRefUse
	if !reflect.DeepEqual(comparison, previous) {
		t.Fatal("field identity, shape, order or exact source spans changed")
	}
}

func TestRecordReferenceRenameDoesNotMaskExplicitTypeEdits(t *testing.T) {
	for _, edit := range []string{"explicit-id", "lookup", "presentation", "removed-target", "duplicate-name"} {
		t.Run(edit, func(t *testing.T) { checkRejectedRecordRename(t, edit) })
	}
}

func checkRejectedRecordRename(t *testing.T, edit string) {
	t.Helper()
	document, model, support := recordRenameFixture(t, "Line")
	editRecordRenameFixture(&model, edit)
	input := model.Clone()
	written, err := PutWithEntityFieldsSupport(document, model, support)
	var putErr *PutError
	if !errors.As(err, &putErr) || !putErr.NoWrite {
		t.Fatal("invalid rename or conflicting type edit must fail without writing", err)
	}
	if !reflect.DeepEqual(written, document) || !reflect.DeepEqual(input, model) {
		t.Fatal("rejected update mutated a source or caller model")
	}
}

func editRecordRenameFixture(model *Model, edit string) {
	renameFixtureRecords(model, false)
	for i := range model.Nodes {
		if edit == "removed-target" && model.Nodes[i].ID == "urn:record-items:line" {
			model.Nodes = append(model.Nodes[:i], model.Nodes[i+1:]...)
			return
		}
		if edit == "duplicate-name" && model.Nodes[i].ID == "urn:record-items:other" {
			model.Nodes[i].Name = "Product"
		}
	}
	for i := range model.Nodes {
		if model.Nodes[i].ID != "urn:record-items:basket" {
			continue
		}
		field := &model.Nodes[i].Fields[0]
		switch edit {
		case "explicit-id":
			field.TypeRef = TypeRef{ID: "urn:record-items:other"}
		case "lookup":
			field.TypeRef = TypeRef{Name: "Other"}
		case "presentation":
			field.TypeRefUse.Spelling = "Other"
		}
	}
}

func TestRecordReferenceRenamePreservesCoherentRetype(t *testing.T) {
	document, model, support := recordRenameFixture(t, "Line")
	for i := range model.Nodes {
		if model.Nodes[i].ID == "urn:record-items:line" {
			model.Nodes[i].Name = "Product"
		}
		if model.Nodes[i].ID == "urn:record-items:basket" {
			field := &model.Nodes[i].Fields[0]
			field.TypeRef = TypeRef{Name: "Other"}
			field.TypeRefUse.Spelling = "Other"
			field.TypeRefUse.ResolvedID = "urn:record-items:other"
		}
	}
	written, err := PutWithEntityFieldsSupport(document, model, support)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := GetWithEntityFieldsSupport(written, support)
	if err != nil || !SemanticEquivalent(model, observed) {
		t.Fatal("coherent explicit retype was overwritten by the rename", err)
	}
	fields := nodeMap(observed.Nodes)["urn:record-items:basket"].Fields
	for i, field := range fields {
		want := ID("urn:record-items:line")
		if i == 0 {
			want = "urn:record-items:other"
		}
		if field.TypeRefUse.ResolvedID != want {
			t.Fatal("explicit retype or unchanged reference was redirected", field)
		}
	}
}

func recordRenameFixture(t *testing.T, spelling string) (Document, Model, EntityFieldsSupport) {
	t.Helper()
	support, _ := syntax.EntityFieldsSupportForProfileID("gooo.entityfields.go-projection.v7")
	source := strings.ReplaceAll(recordReferenceSource, "type Line", "type "+spelling)
	for _, shape := range []string{"optional one", "required one", "optional many"} {
		name := strings.ReplaceAll(shape, " ", "_")
		field := fmt.Sprintf(" field %s id %q type %s %s\n", name, "urn:rename:"+name, spelling, shape)
		source = strings.TrimSuffix(source, "}\n") + field + "}\n"
	}
	source += "entity Other id \"urn:record-items:other\" fields {\n" +
		" field extra id \"urn:record-items:extra\" type integer required one\n}\n"
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("rename.gooo", source, support)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal(err)
	}
	model, err := GetWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	return document, model, support
}

// Reparse the written source spelling without the old resolved-ID cache. This
// detects accidental capture by another record that now owns the old name.
func renamedRecordSource(written Document) string {
	var source strings.Builder
	fmt.Fprintf(&source, "package %s\nnamespace %s\n", written.Package, written.Namespace)
	for _, declaration := range written.Declarations {
		fmt.Fprintf(&source, "entity %s id %q fields {\n", declaration.Name, declaration.ID)
		for _, field := range declaration.Fields {
			spelling := field.TypeRefUse.Spelling
			if field.TypeRefUse.Form == TypeRefFormStableID || strings.Contains(spelling, ":") {
				spelling = fmt.Sprintf("%q", spelling)
			}
			fmt.Fprintf(&source, "field %s id %q type %s %s %s\n",
				field.Name, field.ID, spelling, strings.ToLower(string(field.Presence)),
				strings.ToLower(string(field.Cardinality)))
		}
		source.WriteString("}\n")
	}
	return source.String()
}

// Fresh parsing starts without the old resolved-ID cache.
func assertRenamedRecordSourceReload(t *testing.T, written Document, support EntityFieldsSupport) {
	t.Helper()
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("renamed.gooo", renamedRecordSource(written), support)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := LowerDocumentWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range ir.Graph.Nodes() {
		if node.ID != "urn:record-items:basket" {
			continue
		}
		for _, field := range node.Fields {
			if field.TypeRef.ID != semantic.ID("urn:record-items:line") {
				t.Fatal("fresh source parse changed the stable record target", field.TypeRef)
			}
		}
	}
}
