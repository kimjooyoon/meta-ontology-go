package bodycodegen

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestConditionDeclarationOriginCanonicalIDs2617(t *testing.T) {
	source := []byte(`package origin2617
namespace origin2617
entity Request id "urn:Origin2617:Request" fields {
 field amount id "URN:Origin2617:Amount" type "urn:gooo:type:integer" required one
}
entity Integer id "urn:gooo:type:integer"
activity Check(Request) -> Integer id "URN:Origin2617:Check" computes "if input.amount < 73 { return input.amount + 4 }; return input.amount - 4"
`)
	generated, err := Generate("origin2617.gooo", source, "Check")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := BindBodyConditionSources(source, generated.Report.BodyConditions)
	if err != nil || len(bound) != 1 || bound[0].Status != "MAPPED" || len(bound[0].Fields) != 1 {
		t.Fatal("canonical source identity binding failed", err, bound)
	}
	decl := bound[0].Fields[0].Declaration
	if decl == nil {
		t.Fatal("field declaration unavailable")
	}
	fieldID, fieldErr := semantic.ParseIdentity(decl.StableID)
	activityID, activityErr := semantic.ParseIdentity(bound[0].Declaration.StableID)
	if fieldErr != nil || activityErr != nil || fieldID.String() != "urn:Origin2617:Amount" || !strings.Contains(string(source[decl.Start:decl.End]), "URN:Origin2617:Amount") || activityID.String() != "urn:Origin2617:Check" {
		t.Fatal("authored spelling replaced stable semantic identity", decl)
	}
	t.Log("origin2617: canonical field/activity identities retain original ID spellings and declaration byte ranges; no native execution")
}

func TestConditionDeclarationOriginCanonicalHost2683(t *testing.T) {
	source := []byte(`package origin2683
namespace origin2683
entity Request id "urn:Origin2683:Request" fields {
 field count id "ORIGIN2683://Fields/Count" type "urn:gooo:type:integer" required one
}
entity Integer id "urn:gooo:type:integer"
activity Adjust(Request) -> Integer id "ORIGIN2683://Activities/Adjust" computes "if input.count < 81 { return input.count + 7 }; return input.count - 7"
`)
	generated, err := Generate("origin2683.gooo", source, "Adjust")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := BindBodyConditionSources(source, generated.Report.BodyConditions)
	if err != nil || len(bound) != 1 || bound[0].Status != "MAPPED" || len(bound[0].Fields) != 1 {
		t.Fatal("canonical host identity binding failed", err, bound)
	}
	decl := bound[0].Fields[0].Declaration
	if decl == nil {
		t.Fatal("field declaration unavailable")
	}
	fieldID, fieldErr := semantic.ParseIdentity(decl.StableID)
	activityID, activityErr := semantic.ParseIdentity(bound[0].Declaration.StableID)
	if fieldErr != nil || activityErr != nil || fieldID.String() != "origin2683://fields/Count" || !strings.Contains(string(source[decl.Start:decl.End]), "ORIGIN2683://Fields/Count") || activityID.String() != "origin2683://activities/Adjust" {
		t.Fatal("canonical identity lost authored spelling", decl)
	}
	t.Log("origin2683: source-facing field/activity IDs retain authored declaration bytes and canonical URI identity; static origin only")
}
