package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestLocalInterrupted10337(t *testing.T) {
	const source = `package locals10337
namespace locals10337
entity Integer id "urn:gooo:type:integer"
entity Batch id "urn:locals10337:batch" fields {field values id "urn:locals10337:values" type integer required many}
activity Quotient(Integer,Integer) -> Integer id "urn:locals10337:quotient" computes "let divisor=input1;let result=input0/divisor;return result"
activity Main(Integer,Integer) -> Batch id "urn:locals10337:main" computes "let seed=input0+157;let denominator=input1;if denominator==0 && seed<0 {return Batch{values:[]int64{seed}}};let outputs=[]int64{seed,Quotient(seed,denominator)};return Batch{values:outputs}"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, v := range [][3]string{{`23`, `3`, `{"values":[180,60]}`}, {`-160`, `0`, `{"values":[-3]}`}, {`11`, `0`, `{"values":[168,0]}`}, {`29`, `-2`, `{"values":[186,-93]}`}} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Main.input0": json.RawMessage(v[0]), "Main.input1": json.RawMessage(v[1])}, Expected: map[string]json.RawMessage{"Main": json.RawMessage(v[2])}})
	}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "local-interrupted10337.json", o)
	p, err := GenerateCompositionWithOptions(context.Background(), "interrupted10337.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Main"})
	o["composition"] = p
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "interrupted10337.gooo", []byte(source), p, suite, "")
	o["runtime"] = r
	if err != nil {
		t.Fatal(err)
	}
	s := r.LocalObservation
	if r.FinitePassed != 3 || r.Outcomes.Faulted != 1 || !r.RuntimeReplayed || s == nil || s.Status != "PARTIAL" || s.Interrupted != 1 || s.RecordedWrites != 14 || s.RecordedUses != 21 || s.Frames != 7 || s.ElementWindows != 7 {
		t.Fatal("fault observation", s, r.FinitePassed)
	}
	if err := validateRecordedNativeOrigins(r); err != nil {
		t.Fatal(err)
	}
	changed := r
	changed.LocalSites = append([]CompositionLocalSite(nil), r.LocalSites...)
	changed.LocalSites[0].Name = "other"
	if sameJointRuntime(r, changed) {
		t.Fatal("runtime comparison ignored local site change")
	}
	changed = r
	changed.LocalObservation = new(CompositionLocalSummary)
	*changed.LocalObservation = *r.LocalObservation
	changed.LocalObservation.Interrupted = 0
	if sameJointRuntime(r, changed) || validateRecordedLocalOrigins(changed) == nil {
		t.Fatal("runtime summary mutation accepted")
	}
	t.Log("native3/4; one intended ZERO_DIVISOR;1 interrupted assignment;14 writes;21 uses;7 element windows;changed local source and populations rejected")
}

func TestLocalWriteLimit10357(t *testing.T) {
	const source = `package locals10357
namespace locals10357
entity Integer id "urn:gooo:type:integer"
activity Sum(Integer) -> Integer id "urn:locals10357:sum" computes "let counter=0;let total=input;for counter=0;counter<1024;counter=counter+1 {let item=counter+input;let copy=item;total=total+copy};return total"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, v := range [][2]string{{`2`, `525826`}, {`-3`, `520701`}} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Sum": json.RawMessage(v[0])}, Expected: map[string]json.RawMessage{"Sum": json.RawMessage(v[1])}})
	}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "local-limit10357.json", o)
	p, err := GenerateCompositionWithOptions(context.Background(), "limit10357.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Sum"})
	o["composition"] = p
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "limit10357.gooo", []byte(source), p, suite, "")
	o["runtime"] = r
	if err != nil {
		t.Fatal(err)
	}
	s := r.LocalObservation
	if r.FinitePassed != 2 || !r.RuntimeReplayed || s == nil || s.Status != "PARTIAL" || s.RecordedWrites != 4096 || s.WritesDropped != 4102 || s.RecordedUses != 12292 || s.UsesDropped != 0 || s.Unresolved == 0 {
		t.Fatal("bounded recording", s)
	}
	if err := validateRecordedNativeOrigins(r); err != nil {
		t.Fatal(err)
	}
	t.Logf("native2/2; writes4096;write drops4102;uses12292;unresolved%d; bounded recording preserved results", s.Unresolved)
}
