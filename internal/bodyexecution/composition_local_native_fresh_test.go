package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestLocalNative10267(t *testing.T) {
	const source = `package locals10267
namespace locals10267
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:locals10267:input" fields {
 field values id "urn:locals10267:values" type integer required many
 field flip id "urn:locals10267:flip" type boolean required one
}
entity Batch id "urn:locals10267:batch" fields {field values id "urn:locals10267:output" type integer required many}
activity Choose(Input) -> Integer id "urn:locals10267:choose" computes "if input.flip {let item=input.values[1]+137;return item}else{let item=input.values[0]+149;return item}"
activity Build(Input) -> Batch id "urn:locals10267:build" computes "let total=input.values[0];let first=Choose(input);let second=Choose(input);let name=\"keep\";let enabled=input.flip;if enabled && name==\"keep\" {total=total+first};for entry in input.values {total=total+entry};let results=[]int64{total,second};return Batch{values:results}"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, test := range [][2]string{{`{"values":[3,5],"flip":false}`, `{"values":[11,152]}`}, {`{"values":[3,5],"flip":true}`, `{"values":[153,142]}`}, {`{"values":[7,7],"flip":true}`, `{"values":[165,144]}`}, {`{"values":[-11,13],"flip":false}`, `{"values":[-9,138]}`}} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(test[0])}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(test[1])}})
	}
	observation := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "local-native10267.json", observation)
	program, err := GenerateCompositionWithOptions(context.Background(), "실행기원10267.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	observation["composition"] = program
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "실행기원10267.gooo", []byte(source), program, suite, "")
	observation["runtime"] = runtime
	if err != nil {
		t.Fatal(err)
	}
	if runtime.FinitePassed != 4 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("finite native outcomes", runtime.FinitePassed)
	}
	s := runtime.LocalObservation
	if s == nil || s.Status != "COMPLETE" || s.RecordedWrites != 46 || s.RecordedUses != 42 || s.ElementWindows != 8 || s.Frames != 12 {
		t.Fatal("local population", s)
	}
	for _, trace := range runtime.Traces {
		frames := map[int]bool{}
		iterations := 0
		for _, w := range trace.LocalWrites {
			site := runtime.LocalSites[w.SiteIndex]
			if site.ActivityID == "urn:locals10267:choose" {
				frames[w.Frame] = true
			}
			if site.Kind == "iteration" {
				if w.ReadEnd-w.ReadStart != 1 {
					t.Fatal("iteration source read missing", w)
				}
				read := trace.Reads[w.ReadStart]
				if read.Origin == nil || read.Origin.Index != iterations {
					t.Fatal("iteration input identity", read)
				}
				iterations++
			}
		}
		if len(frames) != 2 || iterations != 2 {
			t.Fatal("call and iteration identities", frames, iterations)
		}
		for _, element := range trace.LocalElements {
			if element.UseEnd-element.UseStart != 1 {
				t.Fatal("constructor scalar use", element)
			}
			use := trace.LocalUses[element.UseStart]
			write := trace.LocalWrites[use.Assignment]
			if element.Element == 0 && runtime.LocalSites[write.SiteIndex].Operation != "=" {
				t.Fatal("constructor did not use latest reassignment", write)
			}
		}
	}
	if err := validateRecordedLocalOrigins(runtime); err != nil {
		t.Fatal(err)
	}
	changed := runtime
	changed.Traces = append([]CompositionTrace(nil), runtime.Traces...)
	changed.Traces[0].LocalUses = append([]CompositionLocalUse(nil), runtime.Traces[0].LocalUses...)
	last := len(changed.Traces[0].LocalUses) - 2
	changed.Traces[0].LocalUses[last].Assignment = 0
	if err := validateRecordedLocalOrigins(changed); err == nil {
		t.Fatal("stale equal-type assignment accepted")
	}
	t.Log("native 4/4;46 writes;42 uses;12 invocation frames;8 constructor windows;stale assignment rejected;model0")
}

func TestScalarOnlyNative10289(t *testing.T) {
	const source = `package locals10289
namespace locals10289
entity Integer id "urn:gooo:type:integer"
activity Step(Integer) -> Integer id "urn:locals10289:step" computes "let start=input+151;let temp=start;start=start+3;return temp+start"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, test := range [][2]string{{`41`, `387`}, {`-151`, `3`}} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Step": json.RawMessage(test[0])}, Expected: map[string]json.RawMessage{"Step": json.RawMessage(test[1])}})
	}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "local-native10289.json", o)
	p, err := GenerateCompositionWithOptions(context.Background(), "scalar10289.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Step"})
	o["composition"] = p
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "scalar10289.gooo", []byte(source), p, suite, "")
	o["runtime"] = r
	if err != nil {
		t.Fatal(err)
	}
	if r.FinitePassed != 2 || !r.RuntimeReplayed || r.LocalObservation == nil || r.LocalObservation.RecordedWrites != 6 || r.LocalObservation.RecordedUses != 8 || len(r.ReadSites) != 0 || len(r.ConstructionSites) != 0 {
		t.Fatal("scalar-only observation", r.LocalObservation)
	}
	t.Log("native2/2;6 writes;8 uses;zero collection reads;model0")
}
