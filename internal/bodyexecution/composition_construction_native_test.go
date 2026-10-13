package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestConstructionOrigins9533(t *testing.T) {
	const source = `package constructing
namespace constructing9533
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:make9533:input" fields {
 field values id "urn:make9533:values" type integer required many
 field pick id "urn:make9533:pick" type integer required one
}
entity Batch id "urn:make9533:batch" fields {
 field values id "urn:make9533:batch-values" type integer required many
 field pick id "urn:make9533:batch-pick" type integer required one
 field labels id "urn:make9533:labels" type string required many
 field flags id "urn:make9533:flags" type boolean required many
 field empty id "urn:make9533:empty" type integer required many
}
activity Inner(Input) -> Integer id "urn:make9533:inner" computes "let values=[]int64{input.values[1]};return values[0]"
activity Build(Input) -> Batch id "urn:make9533:build" computes "let previous=input.values[0];let numbers=[]int64{Inner(input),input.values[0]+input.values[1],previous+2};let labels=[]string{\"start\",\"끝\"};let switches=[]bool{true,false};let empty=[]int64{};return Batch{values:numbers[1:],pick:input.pick,labels:labels,flags:switches,empty:empty}"
activity Read(Batch) -> Integer id "urn:make9533:read" computes "if input.flags[0] && input.labels[1]==\"끝\" {return input.values[input.pick]+int64(len(input.empty))};return -1"
bind Build.result -> Read.input
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range []struct{ input, want string }{
		{`{"values":[541,541],"pick":0}`, `1082`},
		{`{"values":[547,557],"pick":1}`, `549`},
		{`{"values":[557,563],"pick":0}`, `1120`},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(row.input)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(row.want)}})
	}
	observed := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "construction9533.json", observed)
	program, err := GenerateCompositionWithOptions(context.Background(), "목록생성9533.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Read"})
	observed["composition"] = program
	if err != nil {
		t.Fatal("construction", err)
	}
	runtime, err := ExecuteComposition(context.Background(), "목록생성9533.gooo", []byte(source), program, suite, "")
	observed["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 3 || !runtime.RuntimeReplayed {
		t.Fatal("native", err, runtime.FinitePassed)
	}
	summary := runtime.ReadObservation
	if summary == nil || summary.RecordedReads != 24 || summary.ResolvedReads != 12 || summary.ConstructedReads != 12 || summary.UnresolvedReads != 0 || summary.RecordedConstructions != 15 || summary.Status != "COMPLETE" {
		t.Fatal("origin denominators", summary)
	}
	for c, trace := range runtime.Traces {
		if len(trace.Constructions) != 5 || len(trace.Reads) != 8 {
			t.Fatal("case construction count", c, trace)
		}
		outer, inner := trace.Constructions[0], trace.Constructions[1]
		if outer.ReadStart != 1 || outer.ReadEnd != 5 || outer.Elements[0].ReadStart != 1 || outer.Elements[0].ReadEnd != 3 || outer.Elements[1].ReadStart != 3 || outer.Elements[1].ReadEnd != 5 || outer.Elements[2].ReadStart != 5 || outer.Elements[2].ReadEnd != 5 {
			t.Fatal("actual evaluation windows exclude earlier scalar read", outer)
		}
		if inner.ReadStart != 1 || inner.ReadEnd != 2 || trace.Reads[2].Construction == nil || trace.Reads[2].Construction.Instance != 1 {
			t.Fatal("nested construction", inner, trace.Reads[2])
		}
		last := trace.Reads[7]
		want := 1
		if c == 1 {
			want = 2
		}
		if last.Origin != nil || last.Construction == nil || last.Construction.Instance != 0 || last.Construction.Index != want {
			t.Fatal("slice retains construction position", last)
		}
	}
	for _, site := range runtime.ConstructionSites {
		if site.Source == nil || site.Source.Status != "MAPPED" || site.Source.SourceSHA256 != program.SelectedSourceSHA256 {
			t.Fatal("literal location", site)
		}
		for _, e := range site.Elements {
			if e.Source == nil || e.Source.Status != "MAPPED" || e.Source.End <= e.Source.Start {
				t.Fatal("element location", e)
			}
		}
	}
	facts := CompositionConstructionFacts(runtime)
	observed["construction_facts"] = facts
	observed["constructed_read_facts"] = CompositionConstructedReadFacts(runtime)
	if len(facts) != 24 {
		t.Fatal("element facts", len(facts))
	}
	facts[0].Expression = "detached"
	if runtime.ConstructionSites[0].Elements[0].Source.Expression == "detached" {
		t.Fatal("fact alias")
	}
	observed["construction_facts"] = CompositionConstructionFacts(runtime)
	t.Log("three fresh inputs3/3;24 reads12 input+12 constructed;15 instances;24 element facts;scalar prior-local dependency intentionally not inferred")
}

func TestConstructionOnly9547(t *testing.T) {
	const source = `package constructoronly
namespace constructoronly9547
entity Input id "urn:only9547:input" fields {field name id "urn:only9547:name" type string required one}
entity Output id "urn:only9547:output" fields {
 field values id "urn:only9547:values" type integer required many
 field labels id "urn:only9547:labels" type string required many
}
activity Build(Input) -> Output id "urn:only9547:build" computes "return Output{values:[]int64{},labels:[]string{input.name}}"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, name := range []string{"싹", "bud"} {
		raw, _ := json.Marshal(map[string]string{"name": name})
		want, _ := json.Marshal(map[string]any{"values": []int{}, "labels": []string{name}})
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": raw}, Expected: map[string]json.RawMessage{"Build": want}})
	}
	program, err := GenerateComposition(context.Background(), "생성만9547.gooo", []byte(source), suite, "")
	observed := map[string]any{"source": source, "cases": suite, "composition": program}
	defer saveReadObservation(t, "construction9547.json", observed)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "생성만9547.gooo", []byte(source), program, suite, "")
	observed["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 2 || len(runtime.ReadSites) != 0 || runtime.ReadObservation == nil || runtime.ReadObservation.RecordedConstructions != 4 || runtime.ReadObservation.RecordedReads != 0 {
		t.Fatal("constructor-only observation", err, runtime.ReadObservation)
	}
}

func TestConstructionFault9551(t *testing.T) {
	const source = `package interrupted
namespace interrupted9551
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:fault9551:input" fields {
 field values id "urn:fault9551:values" type integer required many
 field divisor id "urn:fault9551:divisor" type integer required one
}
entity Batch id "urn:fault9551:batch" fields {field values id "urn:fault9551:output" type integer required many}
activity Build(Input) -> Batch id "urn:fault9551:build" computes "return Batch{values:[]int64{input.values[0],input.values[1]/input.divisor,23}}"
activity Read(Batch) -> Integer id "urn:fault9551:read" computes "return input.values[1]"
bind Build.result -> Read.input
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[577,587],"divisor":0}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`0`)}},
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[599,601],"divisor":3}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`200`)}},
	}}
	program, err := GenerateComposition(context.Background(), "중단9551.gooo", []byte(source), suite, "")
	observed := map[string]any{"source": source, "cases": suite, "composition": program}
	defer saveReadObservation(t, "construction9551.json", observed)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "중단9551.gooo", []byte(source), program, suite, "")
	observed["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 1 || runtime.ReadObservation == nil || runtime.ReadObservation.InterruptedConstructions != 1 || runtime.ReadObservation.ConstructedReads != 1 {
		t.Fatal("fault semantics", err, runtime.ReadObservation)
	}
	first := runtime.Traces[0].Constructions[0]
	if first.Completed || len(first.Elements) != 2 || !first.Elements[0].Completed || first.Elements[1].Completed || len(runtime.Traces[0].Reads) != 2 {
		t.Fatal("attempted prefix", first)
	}
	t.Log("one successful output;one arithmetic fault with blocked dependent;interrupted constructor keeps only attempted element prefix")
}
