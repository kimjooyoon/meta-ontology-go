package bidir

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestLatentSyntaxFieldsLowerToSemanticIRInDeclarationOrder(t *testing.T) {
	file := latentSyntaxFile()
	support := supportedEntityFieldsForTest()
	document, err := documentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal(err)
	}
	if got := document.Declarations[0].Fields; len(got) != 2 || got[0].ID != "billing://field/order-number" || got[1].ID != "billing://field/amount" {
		t.Fatalf("field order or identity changed: %#v", got)
	}
	if got := document.Declarations[0].Fields; got[0].TypeRef.Name != "string" || got[1].TypeRef.Namespace != "gooo" || got[1].TypeRef.Name != "string" {
		t.Fatalf("type reference lookup spelling changed: %#v", got)
	}

	ir, err := lowerWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal(err)
	}
	order, ok := ir.Graph.Node(semantic.MustIdentity("billing://entity/order"))
	if !ok || len(order.Fields) != 2 {
		t.Fatalf("semantic order fields missing: %#v", order)
	}
	if order.Fields[0].ID != semantic.MustIdentity("billing://field/order-number") || order.Fields[1].ID != semantic.MustIdentity("billing://field/amount") {
		t.Fatalf("semantic field order changed: %#v", order.Fields)
	}
	if order.Fields[0].Parent != order.ID || order.Fields[1].Parent != order.ID {
		t.Fatalf("field parent identity was not explicit: %#v", order.Fields)
	}
	if order.Fields[0].TypeRef.ID != semantic.BuiltinStringTypeID || order.Fields[1].TypeRef.ID != semantic.BuiltinStringTypeID {
		t.Fatalf("registered string type was not resolved: %#v", order.Fields)
	}
	if order.Fields[0].Presence != semantic.Required || order.Fields[1].Presence != semantic.Required || order.Fields[0].Cardinality != semantic.One || order.Fields[1].Cardinality != semantic.One {
		t.Fatalf("field presence/cardinality changed: %#v", order.Fields)
	}
	if order.Fields[0].Span.File != "latent.gooo" || order.Fields[0].Span.Start.Offset != 10 || order.Fields[1].Span.End.Offset != 102 {
		t.Fatalf("field spans were not lowered: %#v", order.Fields)
	}

	payment, ok := ir.Graph.Node(semantic.MustIdentity("billing://entity/payment"))
	if !ok || len(payment.Fields) != 1 || payment.Fields[0].ID != semantic.MustIdentity("billing://field/receipt") {
		t.Fatalf("second entity fields were not lowered: %#v", payment)
	}
}

func TestLatentFieldsGetPutPreservesAliasesSpansAndOrder(t *testing.T) {
	document := latentDocument()
	before := document
	model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Nodes[0].Fields) != 2 || model.Nodes[0].Fields[0].Name != "Order Number" || model.Nodes[0].Fields[1].Name != "Amount" {
		t.Fatalf("model field order changed: %#v", model.Nodes[0].Fields)
	}
	written, err := putWithEntityFieldsSupport(document, model, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatalf("field carrier changed during Get-Put:\n got %#v\nwant %#v", written, document)
	}
	observed, err := getWithEntityFieldsSupport(written, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(observed, model) || !SemanticEquivalent(model, observed) {
		t.Fatalf("field readback changed model:\n got %#v\nwant %#v", observed, model)
	}
	if !reflect.DeepEqual(document, before) {
		t.Fatal("Get-Put mutated the source document")
	}

	clone := model.Clone()
	clone.Nodes[0].Fields[0].Aliases[0] = "changed"
	if model.Nodes[0].Fields[0].Aliases[0] == "changed" || document.Declarations[0].Fields[0].Aliases[0] == "changed" {
		t.Fatal("field aliases were not deep-cloned")
	}
	written.Declarations[0].Fields[0].Aliases[0] = "written-only"
	if model.Nodes[0].Fields[0].Aliases[0] == "written-only" || document.Declarations[0].Fields[0].Aliases[0] == "written-only" {
		t.Fatal("readback field aliases retained caller storage")
	}
}

func TestLatentFieldSemanticIdentityIgnoresPresentationAndSpan(t *testing.T) {
	left, err := getWithEntityFieldsSupport(latentDocument(), supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	right := left.Clone()
	right.Nodes[0].Fields[0].Name = "Renamed presentation"
	right.Nodes[0].Fields[0].Aliases = []string{"new alias"}
	right.Nodes[0].Fields[0].Span = SourceSpan{File: "other.gooo", Start: 100, End: 120}
	right.Nodes[0].Fields[0].TypeRef = semantic.TypeRef{ID: semantic.BuiltinStringTypeID}
	right.Nodes[0].Fields[0].TypeRefUse.Spelling = "renamed:display"
	right.Nodes[0].Fields[0].TypeRefUse.Form = TypeRefFormLookup
	if !SemanticEquivalent(left, right) || SemanticFingerprint(left) != SemanticFingerprint(right) {
		t.Fatal("field presentation or span changed semantic identity")
	}
	right.Nodes[0].Fields[0].Cardinality = FieldCardinalityMany
	if SemanticEquivalent(left, right) || SemanticFingerprint(left) == SemanticFingerprint(right) {
		t.Fatal("field semantic cardinality change was ignored")
	}
}

func TestLatentFieldValidationRejectsDeterministicallyWithoutPartialModel(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Document)
		want   string
	}{
		{name: "duplicate-global-id", mutate: func(document *Document) {
			document.Declarations[1].Fields[0].ID = document.Declarations[0].Fields[0].ID
		}, want: "duplicate field ID"},
		{name: "same-parent-name-alias-collision", mutate: func(document *Document) {
			document.Declarations[0].Fields[1].Name = "legacy-number"
		}, want: "field name"},
		{name: "unknown-type", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].TypeRef = semantic.TypeRef{Name: "missing"}
		}, want: "unknown semantic type"},
		{name: "invalid-presence", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].Presence = FieldPresence("sometimes")
		}, want: "unknown presence"},
		{name: "invalid-cardinality", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].Cardinality = FieldCardinality("unordered")
		}, want: "unknown cardinality"},
		{name: "non-entity-owner", mutate: func(document *Document) {
			document.Declarations[0].Kind = ActivityKind
		}, want: "only valid on Entity"},
		{name: "wrong-parent", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].Parent = "billing://entity/payment"
		}, want: "parent"},
		{name: "invalid-type-ref", mutate: func(document *Document) {
			document.Declarations[0].Fields[0].TypeRef = semantic.TypeRef{ID: "not-an-identity"}
		}, want: "invalid semantic type reference"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			document := latentDocument()
			test.mutate(&document)
			beforeGet := document
			model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Get error = %v, want substring %q", err, test.want)
			}
			if !reflect.DeepEqual(model, Model{}) {
				t.Fatalf("failed Get returned partial model: %#v", model)
			}
			if !reflect.DeepEqual(document, beforeGet) {
				t.Fatal("failed Get mutated the source document")
			}
		})
	}
}

func TestLatentFieldPutRejectsParentMoveWithoutWrite(t *testing.T) {
	document := latentDocument()
	model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	updated := model.Clone()
	field := updated.Nodes[0].Fields[0]
	updated.Nodes[0].Fields = updated.Nodes[0].Fields[1:]
	updated.Nodes[1].Fields = append(updated.Nodes[1].Fields, Field{
		ID: field.ID, Parent: "billing://entity/payment", Name: field.Name, Aliases: field.Aliases,
		TypeRef: field.TypeRef, Presence: field.Presence, Cardinality: field.Cardinality, Span: field.Span,
	})
	written, err := putWithEntityFieldsSupport(document, updated, supportedEntityFieldsForTest())
	if err == nil || !strings.Contains(err.Error(), "cannot move") {
		t.Fatalf("parent move result = %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("parent move changed the source document")
	}
}

func TestLatentFieldPutRejectsSemanticAdditionWithoutFieldSpan(t *testing.T) {
	document := latentDocument()
	document.Declarations[0].Span = SourceSpan{File: "latent.gooo", Start: 1, End: 2}
	model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	updated := model.Clone()
	updated.Nodes[0].Fields = append(updated.Nodes[0].Fields, Field{
		ID: "billing://field/no-span", Parent: "billing://entity/order", Name: "No Span",
		TypeRef: TypeRef{Name: "string"}, Presence: FieldPresenceRequired, Cardinality: FieldCardinalityOne,
	})
	written, err := putWithEntityFieldsSupport(document, updated, supportedEntityFieldsForTest())
	if err == nil || !strings.Contains(err.Error(), "field \"billing://field/no-span\" semantic change has no source span") {
		t.Fatalf("missing field provenance result = %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("missing field provenance changed the source document")
	}
}

func TestLatentFieldPutRejectsWithoutWriteOrCandidatePromotion(t *testing.T) {
	document := latentDocument()
	model, err := getWithEntityFieldsSupport(document, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	updated := model.Clone()
	updated.Candidates = append(updated.Candidates, NewSourcedFact(CandidateFact, "billing://activity/pay", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "candidate.go", Start: 1, End: 2}))
	updated.Nodes[0].Fields[0].TypeRef = semantic.TypeRef{}
	updated.Nodes[0].Fields[0].TypeRefUse = TypeRefUse{}
	written, err := putWithEntityFieldsSupport(document, updated, supportedEntityFieldsForTest())
	if err == nil {
		t.Fatal("invalid field model was accepted")
	}
	putErr, ok := err.(*PutError)
	if !ok || putErr.Code != PutModelInvalid || !putErr.NoWrite {
		t.Fatalf("unexpected transactional Put error: %v", err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatal("rejected Put changed the source document")
	}
	if len(updated.Candidates) != 1 || len(model.Candidates) != 0 {
		t.Fatal("candidate state was unexpectedly promoted or erased")
	}
}

func TestLatentFieldTypedRegistryRejectsUnknownAndAmbiguousReferences(t *testing.T) {
	document := latentDocument()
	ambiguous := semantic.NewTypeRegistry()
	if err := ambiguous.Register(semantic.TypeDef{ID: "urn:custom:type:string", Namespace: "custom", Name: "string"}); err != nil {
		t.Fatal(err)
	}
	document.Declarations[0].Fields[0].TypeRef = semantic.TypeRef{Name: "string"}
	if ir, err := lowerDocumentWithTypesAndEntityFieldsSupport(document, ambiguous, supportedEntityFieldsForTest()); err == nil || !strings.Contains(err.Error(), "ambiguous semantic type") || !reflect.DeepEqual(ir, semantic.IR{}) {
		t.Fatalf("ambiguous type result = %#v, %v", ir, err)
	}
	document.Declarations[0].Fields[0].TypeRef = semantic.TypeRef{Namespace: "gooo", Name: "string"}
	document.Declarations[0].Fields[1].TypeRef = semantic.TypeRef{Namespace: "gooo", Name: "string"}
	document.Declarations[1].Fields[0].TypeRef = semantic.TypeRef{Namespace: "gooo", Name: "string"}
	ir, err := lowerDocumentWithTypesAndEntityFieldsSupport(document, ambiguous, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	field := ir.Graph.Nodes()[0].Fields[0]
	if field.TypeRef.ID != semantic.BuiltinStringTypeID || field.TypeRef.Namespace != "" || field.TypeRef.Name != "" {
		t.Fatalf("typed TypeRef identity did not follow the live semantic canonical form: %#v", field.TypeRef)
	}
}

func TestPublicParserCannotReachLatentFieldsThroughBX(t *testing.T) {
	file, diagnostics := syntax.Parse(`package billing
namespace billing
entity Order id "billing://entity/order" field Name id "billing://field/name" type string required one`)
	if len(diagnostics) == 0 || diagnostics[0].Code != syntax.DiagUnexpectedDeclaration {
		t.Fatalf("proposed public field syntax was accepted: %#v", diagnostics)
	}
	if file == nil || len(file.Declarations) == 0 || len(file.Declarations[0].(*syntax.EntityDecl).Fields) != 0 {
		t.Fatal("public parser produced a partial latent field AST")
	}
}

func latentSyntaxFile() *syntax.File {
	return &syntax.File{
		Package:   &syntax.PackageDecl{Name: "billing"},
		Namespace: &syntax.NamespaceDecl{Name: "billing"},
		Declarations: []syntax.Declaration{
			&syntax.EntityDecl{
				Name: "Order", ID: "billing://entity/order",
				Fields: []syntax.FieldDecl{
					{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 10}, End: syntax.Position{Offset: 50}}, ID: "billing://field/order-number", Name: "Order Number", TypeRef: syntax.TypeRefDecl{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 31}, End: syntax.Position{Offset: 37}}, Spelling: "string"}, Presence: syntax.FieldPresenceRequired, Cardinality: syntax.FieldCardinalityOne,
						IDSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 10}, End: syntax.Position{Offset: 18}}, NameSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 19}, End: syntax.Position{Offset: 30}}, PresenceSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 38}, End: syntax.Position{Offset: 46}}, CardinalitySpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 47}, End: syntax.Position{Offset: 50}}},
					{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 60}, End: syntax.Position{Offset: 102}}, ID: "billing://field/amount", Name: "Amount", TypeRef: syntax.TypeRefDecl{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 81}, End: syntax.Position{Offset: 92}}, Spelling: "gooo:string"}, Presence: syntax.FieldPresenceRequired, Cardinality: syntax.FieldCardinalityOne,
						IDSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 60}, End: syntax.Position{Offset: 68}}, NameSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 69}, End: syntax.Position{Offset: 80}}, PresenceSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 93}, End: syntax.Position{Offset: 97}}, CardinalitySpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 98}, End: syntax.Position{Offset: 102}}},
				},
			},
			&syntax.EntityDecl{
				Name: "Payment", ID: "billing://entity/payment",
				Fields: []syntax.FieldDecl{{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 110}, End: syntax.Position{Offset: 150}}, ID: "billing://field/receipt", Name: "Amount", TypeRef: syntax.TypeRefDecl{Span: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 131}, End: syntax.Position{Offset: 137}}, Spelling: "string"}, Presence: syntax.FieldPresenceRequired, Cardinality: syntax.FieldCardinalityOne,
					IDSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 110}, End: syntax.Position{Offset: 118}}, NameSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 119}, End: syntax.Position{Offset: 130}}, PresenceSpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 138}, End: syntax.Position{Offset: 146}}, CardinalitySpan: syntax.Span{Filename: "latent.gooo", Start: syntax.Position{Offset: 147}, End: syntax.Position{Offset: 150}}}},
			},
		},
	}
}
