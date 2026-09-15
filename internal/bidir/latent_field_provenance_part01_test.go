package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestLatentFieldSourceTypeRefUseSurvivesRegistryRename(t *testing.T) {
	oldRegistry := semantic.NewTypeRegistry()
	const typeID semantic.ID = semantic.BuiltinStringTypeID
	document := customTypeDocument(TypeRef{ID: typeID}, TypeRefUse{Form: TypeRefFormLookup, Spelling: "string", ResolvedID: ID(typeID), Span: SourceSpan{File: "custom.gooo", Start: 26, End: 32}})
	support := supportedEntityFieldsForTest()
	model, err := getWithTypesAndEntityFieldsSupport(document, oldRegistry, support)
	if err != nil {
		t.Fatal(err)
	}
	firstWrite, err := putWithTypesAndEntityFieldsSupport(document, model, oldRegistry, support)
	if err != nil {
		t.Fatal(err)
	}

	renamedRegistry := semantic.NewTypeRegistry()
	written, err := putWithTypesAndEntityFieldsSupport(firstWrite, model, renamedRegistry, support)
	if err != nil {
		t.Fatal(err)
	}
	field := written.Declarations[0].Fields[0]
	if field.TypeRefUse.Form != TypeRefFormLookup || field.TypeRefUse.Spelling != "string" || field.TypeRefUse.ResolvedID != ID(typeID) {
		t.Fatalf("registry rename rewrote source TypeRefUse: %#v", field.TypeRefUse)
	}

	stableDocument := customTypeDocument(TypeRef{ID: typeID}, TypeRefUse{Form: TypeRefFormStableID, Spelling: string(typeID), ResolvedID: ID(typeID), Span: SourceSpan{File: "custom.gooo", Start: 26, End: 35}})
	stableModel, err := getWithTypesAndEntityFieldsSupport(stableDocument, renamedRegistry, support)
	if err != nil {
		t.Fatal(err)
	}
	stableWritten, err := putWithTypesAndEntityFieldsSupport(stableDocument, stableModel, renamedRegistry, support)
	if err != nil {
		t.Fatal(err)
	}
	stableUse := stableWritten.Declarations[0].Fields[0].TypeRefUse
	if stableUse.Form != TypeRefFormStableID || stableUse.Spelling != string(typeID) || stableUse.ResolvedID != ID(typeID) {
		t.Fatalf("stable-ID source TypeRefUse changed: %#v", stableUse)
	}
}
