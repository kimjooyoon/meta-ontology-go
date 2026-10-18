package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const nativeFilterSource = `package filtering
namespace native_filtering
entity Text id "urn:gooo:type:string"
entity Boolean id "urn:gooo:type:boolean"
entity Batch id "urn:native-filter:batch" fields {
 field values id "urn:native-filter:values" type string required many
}
activity Retain(Text) -> Boolean id "urn:native-filter:retain" computes "return len(input)>0"
activity Build(Batch) -> Batch id "urn:native-filter:build" computes "return Batch{values:filter_values(input.values,Retain)}"
activity Read(Batch) -> Text id "urn:native-filter:read" computes "let text=\"\";for _, value := range input.values {text=text+value};return text"
bind Build.result -> Read.input
`

func TestCollectionFilterNativeOrigins(t *testing.T) {
	for _, sliced := range []bool{false, true} {
		name, source := "whole", nativeFilterSource
		if sliced {
			name = "slice"
			source = strings.Replace(source, "filter_values(input.values,Retain)", "filter_values(input.values[1:],Retain)", 1)
		}
		t.Run(name, func(t *testing.T) {
			rows := [][]string{{"", "same", "same", "한글", ""}, {"", ""}, {}}
			if sliced {
				rows[2] = []string{"skip"}
			}
			suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
			selectedIndexes := make([][]int, len(rows))
			for c, values := range rows {
				input, _ := json.Marshal(map[string]any{"values": values})
				offset := 0
				if sliced {
					offset = 1
				}
				retained := []string{}
				for i, value := range values[offset:] {
					if value != "" {
						retained = append(retained, value)
						selectedIndexes[c] = append(selectedIndexes[c], i)
					}
				}
				output, _ := json.Marshal(map[string]any{"values": retained})
				text, _ := json.Marshal(strings.Join(retained, ""))
				suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": input}, Expected: map[string]json.RawMessage{"Build": output, "Read": text}})
			}
			program, err := GenerateCompositionWithOptions(context.Background(), "filter.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Read"})
			if err != nil {
				t.Fatal(err)
			}
			runtime, err := ExecuteComposition(context.Background(), "filter.gooo", []byte(source), program, suite, "")
			if err != nil || runtime.FinitePassed != 2*len(rows) || !runtime.RuntimeReplayed {
				t.Fatal(err, runtime.Failure, runtime.FinitePassed, runtime.FiniteTotal, runtime.RuntimeReplayed)
			}
			if len(runtime.ConstructionSites) != 1 || runtime.ConstructionSites[0].Kind != "filter_values" || runtime.ConstructionSites[0].PredicateActivityID != "urn:native-filter:retain" || runtime.ConstructionSites[0].Source.Status != "MAPPED" {
				t.Fatal("filter source identity", runtime.ConstructionSites)
			}
			if len(runtime.ConditionSites) != 0 || runtime.ReadObservation == nil || runtime.ReadObservation.Status != "COMPLETE" || runtime.ModelCalls != 0 {
				t.Fatal("generated filter branch must not be a source condition", runtime.ConditionSites, runtime.ReadObservation)
			}
			for c, trace := range runtime.Traces {
				if len(trace.Constructions) != 1 || trace.Constructions[0].ElementCount == nil || *trace.Constructions[0].ElementCount != len(selectedIndexes[c]) {
					t.Fatal("selected count", c, trace.Constructions)
				}
				created := trace.Constructions[0]
				inputCount := len(rows[c])
				offset := 0
				if sliced {
					inputCount--
					offset = 1
				}
				if created.ReadEnd-created.ReadStart != inputCount || len(created.Elements) != len(selectedIndexes[c]) || len(trace.Reads) != inputCount+len(selectedIndexes[c]) {
					t.Fatal("input and output reads", c, trace)
				}
				for i := 0; i < inputCount; i++ {
					read := trace.Reads[i]
					if read.Origin == nil || read.Origin.Index != i+offset {
						t.Fatal("original input position", c, i, read)
					}
				}
				for i, inputIndex := range selectedIndexes[c] {
					element, read := created.Elements[i], trace.Reads[inputCount+i]
					if element.InputIndex == nil || *element.InputIndex != inputIndex || element.ReadStart != inputIndex || element.ReadEnd != inputIndex+1 || read.Construction == nil || read.Construction.Index != i {
						t.Fatal("retained input lineage", c, i, element, read)
					}
				}
			}
			if err := validateRecordedNativeOrigins(runtime); err != nil {
				t.Fatal("recorded filter origin", err)
			}
			if len(selectedIndexes[0]) > 0 {
				bad := -1
				runtime.Traces[0].Constructions[0].Elements[0].InputIndex = &bad
				if validateRecordedNativeOrigins(runtime) == nil {
					t.Fatal("accepted corrupted selection origin")
				}
			}
		})
	}
}

func TestCollectionFilterNestedInputOrigins(t *testing.T) {
	source := strings.Replace(nativeFilterSource, "filter_values(input.values,Retain)", "filter_values(filter_values(input.values,Retain),Retain)", 1)
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":["","same","한글"]}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`"same한글"`)}}}}
	program, err := GenerateCompositionWithOptions(context.Background(), "nested-filter.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Read"})
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "nested-filter.gooo", []byte(source), program, suite, "")
	if err != nil || runtime.FinitePassed != 1 || !runtime.RuntimeReplayed {
		t.Fatal(err, runtime.Failure, runtime.FinitePassed)
	}
	if runtime.ReadObservation.ResolvedReads != 3 || runtime.ReadObservation.ConstructedReads != 4 || runtime.ReadObservation.RecordedConstructions != 2 {
		t.Fatal("inner and outer origins", runtime.ReadObservation)
	}
	trace := runtime.Traces[0]
	if len(trace.Constructions) != 2 || trace.Constructions[0].Elements[0].InputIndex == nil || *trace.Constructions[0].Elements[0].InputIndex != 1 || trace.Constructions[1].Elements[0].InputIndex == nil || *trace.Constructions[1].Elements[0].InputIndex != 0 {
		t.Fatal("each filter retains its own input positions", trace.Constructions)
	}
	if err := validateRecordedNativeOrigins(runtime); err != nil {
		t.Fatal("nested filter origin validation", err)
	}
}
