package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalUseCapacity10543(t *testing.T) {
	const source = `package locals10543
namespace locals10543
entity Integer id "urn:gooo:type:integer"
activity Sum(Integer) -> Integer id "urn:locals10543:sum" computes "let counter=0;let total=input;let origin=input+191;for counter=0;counter<1024;counter=counter+1 {total=total+origin+counter};return total"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for input := -7; input < 7; input++ {
		raw, _ := json.Marshal(input)
		expected, _ := json.Marshal(719360 + 1025*input)
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Sum": raw}, Expected: map[string]json.RawMessage{"Sum": expected}})
	}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "local-use-cap10543.json", o)
	p, err := GenerateCompositionWithOptions(context.Background(), "use-cap10543.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Sum"})
	o["composition"] = p
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "use-cap10543.gooo", []byte(source), p, suite, "")
	if err != nil {
		o["runtime"] = r
		t.Fatal(err)
	}
	s := r.LocalObservation
	// Keep a compact observation; the protocol test itself checks every retained row.
	o["summary"], o["finite_passed"], o["runtime_replayed"], o["native_runs"] = s, r.FinitePassed, r.RuntimeReplayed, r.Runs
	if r.FinitePassed != 14 || !r.RuntimeReplayed || s == nil || s.Status != "PARTIAL" || s.RecordedUses != 65536 || s.UsesDropped != 6172 || s.RecordedWrites != 4096 || s.WritesDropped != 24632 || s.Frames != 14 {
		t.Fatal("use recording bound", s, r.FinitePassed)
	}
	if err := validateRecordedNativeOrigins(r); err != nil {
		t.Fatal(err)
	}
	rows, page, err := SelectRecordedNativeOrigins(r, NativeOriginQuery{Kind: "local-uses", Offset: 65534, Limit: 2})
	if err != nil || len(rows) != 2 || page.TotalRows != 65536 || page.NextOffset != nil || page.LocalObservation.UsesDropped != 6172 {
		t.Fatal("partial final page", err, page)
	}
	t.Log("Gooo native14/14;65536 uses retained,6172 dropped;4096 writes retained,24632 dropped;final page keeps PARTIAL")
}

func TestNativeLocalFrameCapacity10567(t *testing.T) {
	// Exercise the exact generated observer independently of the language's
	// lower helper-call budget. This is a backend protocol boundary observation.
	const harness = `package main
import("encoding/json";"os")
var GoooReadCase,GoooReadCases,GoooReadCaseStart,GoooReadConstructionStart int
var GoooReadEvents []int
var GoooReadDropped [128]int
func main(){
 GoooReadCases=1;GoooReadLocalReset();var total int64
 for i:=range 65537 {frame:=GoooReadLocalFrame();previous:=-1;value:=GoooReadLocalAssign(0,frame,&previous,func()int64{return int64(i)+197});total+=GoooReadLocalUse(0,frame,previous,value)}
 if total!=int64(65537)*65536/2+197*65537 {panic("recording changed arithmetic")}
 GoooReadLocalSave()
}
`
	root := t.TempDir()
	filename := filepath.Join(root, "main.go")
	if err := os.WriteFile(filename, []byte(harness+localObserver), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, processObservation, err := process(ctx, root, nativeTool(), nil, "run", filename)
	if err != nil {
		t.Fatal("new native observer boundary harness", err, processObservation)
	}
	r := CompositionRuntime{LocalSites: []CompositionLocalSite{{RootActivityID: "urn:local10567:root", ActivityID: "urn:local10567:root", VariableOrdinal: 0, Name: "value", ScalarType: "int64", Start: 0, End: 1}}, LocalUseSites: []CompositionLocalUseSite{{RootActivityID: "urn:local10567:root", ActivityID: "urn:local10567:root", VariableOrdinal: 0, Name: "value", ScalarType: "int64", Start: 0, End: 1}}}
	var run compositionLocalRun
	if err := run.observe(0, root, &r); err != nil {
		t.Fatal(err)
	}
	traces := []CompositionTrace{{CaseIndex: 0, Deliveries: []CompositionDelivery{{ActivityID: "urn:local10567:root"}}}}
	if err := run.bind(compositionGraph{}, traces, &r); err != nil {
		t.Fatal(err)
	}
	s := r.LocalObservation
	if s == nil || s.Status != "PARTIAL" || s.Frames != 65536 || s.FramesDropped != 1 || s.RecordedWrites != 4096 || s.WritesDropped != 61441 || s.RecordedUses != 65536 || s.UsesDropped != 1 || s.Unresolved != 61440 {
		t.Fatal("frame capacity", s)
	}
	saveReadObservation(t, "local-frame-cap10567.json", map[string]any{"scope": "generated native observer backend; no Gooo program executed", "harness": harness, "summary": s, "process": processObservation})
	t.Log("backend observer:65536 frames,1 dropped;65537 arithmetic evaluations preserved;bounded arrays;no Gooo language execution claim")
}
