package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"strings"
	"testing"
)

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
