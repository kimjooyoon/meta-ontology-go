package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestScalarLocalBindings10163(t *testing.T) {
	const source = `package locals10163
namespace locals10163
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:locals10163:input" fields {
 field values id "urn:locals10163:values" type integer required many
 field flip id "urn:locals10163:flip" type boolean required one
}
entity Batch id "urn:locals10163:batch" fields { field values id "urn:locals10163:out-values" type integer required many }
activity Choose(Input) -> Integer id "urn:locals10163:choose" computes "if input.flip {let item=input.values[1]+127;return item}else{let item=input.values[0]+131;return item}"
activity Build(Input) -> Batch id "urn:locals10163:build" computes "let 초기=Choose(input);let copied=초기;if input.flip {초기=input.values[0]};let label=\"나\";let toggle=input.flip;for entry in input.values {copied=copied+entry};let items=[]int64{초기,copied};return Batch{values:items}"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[719,727],"flip":true}`)}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[719,2300]}`)}},
	}}
	composition, err := GenerateCompositionWithOptions(context.Background(), "지역기원10163.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	if err != nil {
		t.Fatal(err)
	}
	sites, err := compositionLocalSites(composition)
	if err != nil {
		t.Fatal(err)
	}
	saveReadObservation(t, "local-sites10163.json", map[string]any{"source": source,
		"composition": composition, "sites": sites, "native_executions": 0})
	if len(sites) != 9 {
		t.Fatal("scalar write population", len(sites), sites)
	}
	var sameBinding []CompositionLocalSite
	var siblings []CompositionLocalSite
	types := map[string]int{}
	for _, site := range sites {
		types[site.ScalarType]++
		if site.Source.Value.Status != "MAPPED" || site.Source.Declaration.Status != "MAPPED" || site.RootActivityID != "urn:locals10163:build" || site.ExpressionID == "" {
			t.Fatal("source identity", site)
		}
		if site.Name == "초기" {
			sameBinding = append(sameBinding, site)
		}
		if site.Name == "item" {
			siblings = append(siblings, site)
		}
		if site.Name == "items" {
			t.Fatal("collection binding was claimed as scalar")
		}
	}
	if types["int64"] != 7 || types["string"] != 1 || types["bool"] != 1 || len(sameBinding) != 2 || len(siblings) != 2 {
		t.Fatal("type and lexical populations", types, sameBinding, siblings)
	}
	if sameBinding[0].VariableOrdinal != sameBinding[1].VariableOrdinal || siblings[0].VariableOrdinal == siblings[1].VariableOrdinal {
		t.Fatal("reassignment or sibling identity merged")
	}
	changed := composition
	changed.Steps = append([]CompositionStep(nil), composition.Steps...)
	origins := append([]bodycodegen.BodyLocalOrigin(nil), changed.Steps[0].Generation.Report.BodyLocals...)
	origins[0].Name = "unrelated"
	changed.Steps[0].Generation.Report.BodyLocals = origins
	if _, err := compositionLocalSites(changed); err == nil {
		t.Fatal("mismatching retained write accepted")
	}
	t.Log("9 scalar write sites; 3 scalar types; sibling declarations and Korean reassignment remain distinct; no native execution")
}
