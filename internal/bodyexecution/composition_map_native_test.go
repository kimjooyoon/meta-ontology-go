package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const nativeMapSource = `package mapping
namespace native_mapping
entity Text id "urn:gooo:type:string"
entity Batch id "urn:native-map:batch" fields {
 field values id "urn:native-map:values" type string required many
}
activity Quote(Text) -> Text id "urn:native-map:quote" computes "let quoted=\"[\"+input+\"]\";return quoted"
activity Build(Batch) -> Batch id "urn:native-map:build" computes "let mapped=map_values(input.values, Quote);let size=int64(len(mapped));if size==0{return Batch{values:mapped}};return Batch{values:mapped}"
activity Read(Batch) -> Text id "urn:native-map:read" computes "let text=\"\";for _, value := range input.values {text=text+value};return text"
bind Build.result -> Read.input
`

func TestCollectionMapNativeOrigins(t *testing.T) {
	for _, sliced := range []bool{false, true} {
		name := "whole"
		source := nativeMapSource
		if sliced {
			name = "slice"
			source = strings.Replace(source, "map_values(input.values, Quote)", "map_values(input.values[1:], Quote)", 1)
		}
		t.Run(name, func(t *testing.T) {
			rows := [][]string{{}, {"same", "same", "한글"}, make([]string, 256)}
			for i := range rows[2] {
				rows[2][i] = "x"
			}
			if sliced {
				rows[0] = []string{"skip"}
			}
			suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
			for _, values := range rows {
				input, _ := json.Marshal(map[string]any{"values": values})
				selected := values
				if sliced {
					selected = values[1:]
				}
				mapped := make([]string, len(selected))
				for i, v := range selected {
					mapped[i] = "[" + v + "]"
				}
				output, _ := json.Marshal(map[string]any{"values": mapped})
				text, _ := json.Marshal(strings.Join(mapped, ""))
				suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": input}, Expected: map[string]json.RawMessage{"Build": output, "Read": text}})
			}
			program, err := GenerateCompositionWithOptions(context.Background(), "map.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Read"})
			if err != nil {
				t.Fatal(err)
			}
			runtime, err := ExecuteComposition(context.Background(), "map.gooo", []byte(source), program, suite, "")
			if err != nil || runtime.FinitePassed != 6 || !runtime.RuntimeReplayed {
				t.Fatal(err, runtime.Failure, runtime.FinitePassed, runtime.FiniteTotal, runtime.RuntimeReplayed)
			}
			if len(runtime.ConstructionSites) != 1 {
				t.Fatal("constructor inventory", runtime.ConstructionSites)
			}
			site := runtime.ConstructionSites[0]
			if site.Kind != "map_values" || site.MapperActivityID != "urn:native-map:quote" || site.Source.Status != "MAPPED" || len(site.Elements) != 1 || site.Elements[0].Source.Expression != "Quote" {
				t.Fatal("map source", site)
			}
			if runtime.ReadObservation.Status != "COMPLETE" || runtime.ModelCalls != 0 {
				t.Fatal(runtime.ReadObservation)
			}
			for c, trace := range runtime.Traces {
				count := len(rows[c])
				offset := 0
				if sliced {
					count--
					offset = 1
				}
				if len(trace.Constructions) != 1 || trace.Constructions[0].ElementCount == nil || *trace.Constructions[0].ElementCount != count || len(trace.Reads) != 2*count {
					t.Fatal("population", c, trace.Constructions, len(trace.Reads))
				}
				for i := 0; i < count; i++ {
					input, output := trace.Reads[i], trace.Reads[count+i]
					if input.Origin == nil || input.Origin.Index != i+offset || output.Construction == nil || output.Construction.Instance != 0 || output.Construction.Index != i {
						t.Fatal("position lineage", c, i, input, output)
					}
					part := trace.Constructions[0].Elements[i]
					if !part.Completed || part.ReadStart != i || part.ReadEnd != i+1 {
						t.Fatal("element evaluation", part)
					}
				}
			}
			if err := validateRecordedNativeOrigins(runtime); err != nil {
				t.Fatal("recorded validation", err)
			}
			facts := CompositionConstructionFacts(runtime)
			want := 259
			if sliced {
				want = 257
			}
			if len(facts) != want {
				t.Fatal("facts", len(facts), want)
			}
			for _, fact := range facts {
				if fact.Mapping != "MAPPED" || fact.Expression != "Quote" {
					t.Fatal("source template", fact)
				}
			}
			count := *runtime.Traces[1].Constructions[0].ElementCount
			bad := count + 1
			runtime.Traces[1].Constructions[0].ElementCount = &bad
			if validateRecordedNativeOrigins(runtime) == nil {
				t.Fatal("accepted changed map cardinality")
			}
		})
	}
}
