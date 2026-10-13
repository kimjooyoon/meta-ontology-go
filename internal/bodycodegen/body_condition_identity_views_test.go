package bodycodegen

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestConditionDeclarationSourceAndSemanticViews2791(t *testing.T) {
	source := []byte(`package origin2791
namespace origin2791
entity Request id "urn:Origin2791:Request" fields {
 field amount id "origin2791://Field/Amount" type "urn:gooo:type:integer" required one
 field flag id "URN:Origin2791:Flag" type "urn:gooo:type:boolean" required one
}
entity Integer id "urn:gooo:type:integer"
activity Assess(Request) -> Integer id "origin2791://Activity/Assess" computes "if input.flag && input.amount >= 89 { return input.amount * 2 }; return input.amount / 2"
`)
	generated, err := Generate("origin2791.gooo", source, "Assess")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := BindBodyConditionSources(source, generated.Report.BodyConditions)
	if err != nil || len(bound) != 1 || bound[0].Status != "MAPPED" || len(bound[0].Fields) != 2 {
		t.Fatal("source identity views failed", err, bound)
	}
	expected := map[string]string{"amount": "origin2791://field/Amount", "flag": "urn:Origin2791:Flag"}
	for _, field := range bound[0].Fields {
		decl := field.Declaration
		if decl == nil || decl.StableID != field.FieldID || !strings.Contains(string(source[decl.Start:decl.End]), field.FieldID) {
			t.Fatal("declaration lost the source-facing field ID", field)
		}
		id, err := semantic.ParseIdentity(decl.StableID)
		if err != nil || id.String() != expected[field.Name] {
			t.Fatal("semantic URI identity differs", field, err)
		}
	}
	t.Log("origin2791: source-facing IDs retained2/2 alongside URI-normalized identities; authored field declaration bytes2/2; static only")
}
