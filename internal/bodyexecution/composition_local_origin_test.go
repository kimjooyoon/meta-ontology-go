package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestLocalOriginComposition9911(t *testing.T) {
	const source = `package locals9911
namespace locals9911
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:locals9911:input" fields {
 field values id "urn:locals9911:values" type integer required many
 field reverse id "urn:locals9911:reverse" type boolean required one
}
entity Batch id "urn:locals9911:batch" fields {field values id "urn:locals9911:out-values" type integer required many}
activity Build(Input) -> Batch id "urn:locals9911:build" computes "let original=input.values[0];if input.reverse {original=input.values[1]};let next=original+67;let results=[]int64{next,original};return Batch{values:results}"
activity Read(Batch) -> Integer id "urn:locals9911:read" computes "return input.values[0]-input.values[1]"
bind Build.result -> Read.input
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, raw := range []string{`{"values":[607,613],"reverse":false}`, `{"values":[617,619],"reverse":true}`, `{"values":[631,631],"reverse":true}`} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(raw)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`67`)}})
	}
	observed := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "local9911.json", observed)
	composition, err := GenerateCompositionWithOptions(context.Background(), "지역구성9911.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Read"})
	observed["composition"] = composition
	if err != nil {
		t.Fatal(err)
	}
	origins := composition.Steps[0].Generation.Report.BodyLocals
	if len(origins) != 4 || origins[0].VariableOrdinal != origins[1].VariableOrdinal {
		t.Fatal("source write identity missing", origins)
	}
	bound, err := bodycodegen.BindBodyLocalSources([]byte(composition.GoooSource), origins)
	observed["local_sources"] = bound
	if err != nil || len(bound) != 4 {
		t.Fatal(err)
	}
	for _, location := range bound {
		if location.Value.Status != "MAPPED" || location.Declaration.Status != "MAPPED" {
			t.Fatal("selected source scope", location)
		}
	}
	runtime, err := ExecuteComposition(context.Background(), "지역구성9911.gooo", []byte(source), composition, suite, "")
	observed["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 3 || !runtime.RuntimeReplayed {
		t.Fatal("new source native behavior", err, runtime.FinitePassed)
	}
	for _, trace := range runtime.Traces {
		if len(trace.Constructions) != 1 {
			t.Fatal("construction population")
		}
		for _, element := range trace.Constructions[0].Elements {
			if element.ReadStart != element.ReadEnd {
				t.Fatal("local identity was mislabeled as a direct evaluation read")
			}
		}
	}
	changed := composition
	changed.Steps = append([]CompositionStep(nil), composition.Steps...)
	changed.Steps[0].Generation.Report.BodyLocals = append([]bodycodegen.BodyLocalOrigin(nil), origins...)
	changed.Steps[0].Generation.Report.BodyLocals[1].VariableOrdinal++
	if _, err := replayComposition(context.Background(), "changed9911.gooo", []byte(source), changed); err == nil {
		t.Fatal("changed local binding accepted in source replay")
	}
	t.Log("new source native3/3; four authored writes bound; changed variable identity rejected; element input-read window remains0; scalar origins are tracked separately")
}
