package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestNativeConditionEvidenceKeepsShortCircuitAndFaults941(t *testing.T) {
	source := []byte(`package routes941
namespace routes941
entity Integer id "routes941://integer"
entity Boolean id "routes941://boolean"
activity Eligible(Integer,Integer) -> Boolean id "routes941://eligible" computes "if input0 <= 0 { return false }; if input0 / input1 > 2 { return true }; return false"
activity Main(Integer,Integer) -> Integer id "routes941://main" computes "let GoooObservedCondition = input0; if input0 > 0 && Eligible(input0,input1) { return 91 }; return GoooObservedCondition"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Main.input0":-5,"Main.input1":0},"expected":{"Main":-5}},
{"inputs":{"Main.input0":8,"Main.input1":2},"expected":{"Main":91}},
{"inputs":{"Main.input0":4,"Main.input1":2},"expected":{"Main":4}},
{"inputs":{"Main.input0":0,"Main.input1":0},"expected":{"Main":0}},
{"inputs":{"Main.input0":9,"Main.input1":0},"expected":{"Main":9}}
]}`)
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "routes941.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	original := compositionDigest(prior)
	run, err := ExecuteComposition(ctx, "routes941.gooo", source, prior, suite, nativeTool())
	if err != nil || !run.RuntimeReplayed || len(run.ConditionSites) != 3 || run.FinitePassed != 4 ||
		run.FiniteTotal != 5 || run.Outcomes == nil || run.Outcomes.Faulted != 1 || run.ModelCalls != 0 {
		t.Fatal("actual conditional calls or arithmetic outcomes lost", err, run.Stage)
	}
	want := [][]bool{{false}, {false, true, true}, {false, false, false}, {false}, {false}}
	for c, values := range want {
		trace := run.Traces[c]
		if !trace.ConditionsObserved || len(trace.Conditions) != len(values) {
			t.Fatalf("case %d invented a skipped, repeated or faulted condition: %+v", c, trace.Conditions)
		}
		for i, value := range values {
			if trace.Conditions[i].Value != value {
				t.Fatalf("case %d condition %d: %+v", c, i, trace.Conditions[i])
			}
		}
	}
	for _, site := range run.ConditionSites {
		if site.ProjectionSHA256 != prior.GeneratedSHA256 || site.ExpressionID == "" || site.RootActivityID != "routes941://main" ||
			prior.Source[site.Start:site.End] != site.Expression ||
			(site.ActivityID != "routes941://main" && site.ActivityID != "routes941://eligible") {
			t.Fatal("condition evidence lost its pure projection or declared identity", site)
		}
	}
	if run.Traces[1].Conditions[2].SiteIndex != run.Traces[0].Conditions[0].SiteIndex ||
		run.ConditionSites[run.Traces[1].Conditions[0].SiteIndex].ActivityID != "routes941://eligible" ||
		len(run.Traces[4].Deliveries[0].Fault.Site.Expression) == 0 || compositionDigest(prior) != original ||
		VerifyComposition(ctx, "routes941.gooo", source, prior) != nil {
		t.Fatal("condition recording changed source or helper evaluation order")
	}
	changed := run
	changed.ConditionSites = append([]CompositionConditionSite(nil), run.ConditionSites...)
	changed.ConditionSites[0].Expression += "changed"
	if sameJointRuntime(run, changed) {
		t.Fatal("modified condition source evidence accepted")
	}
	saveConditionObservation(t, "routes941", source, run)
	t.Log("routes941: native 4/5, fault1/5, evaluated conditions9; short-circuit skipped helpers2; native runs agree; source unchanged; model0")
}

func TestNativeConditionEvidenceKeepsSelectedRecordReasons973(t *testing.T) {
	source := []byte(`package checks973
namespace checks973
entity Request id "checks973://request" fields {
 field quantity id "checks973://quantity" type integer required one
 field limit id "checks973://limit" type integer required one
}
entity Decision id "checks973://decision" fields {
 field approved id "checks973://approved" type boolean required one
 field reason id "checks973://reason" type string required one
}
activity Decide(Request) -> Decision id "checks973://decide" computes "if input.quantity < 1 { return Decision{approved:false, reason:\"invalid\"} }; if input.quantity > input.limit { return Decision{approved:false, reason:\"limit\"} }; return __GOOO_BODY_HOLE_result__" assembling {
 source_fill intent "Keep approval and its declared explanation together." {
  hole "result"
  candidate "allow" { fill "result" "Decision{approved:true, reason:\"available\"}" }
  candidate "mislabel" { fill "result" "Decision{approved:true, reason:\"unknown\"}" }
 }
 value_case "[{\"quantity\":3,\"limit\":9}]" -> "{\"approved\":true,\"reason\":\"available\"}"
 holdout_value_case "[{\"quantity\":6,\"limit\":10}]" -> "{\"approved\":true,\"reason\":\"available\"}"
}
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Decide":{"quantity":-3,"limit":7}},"expected":{"Decide":{"approved":false,"reason":"invalid"}}},
{"inputs":{"Decide":{"quantity":13,"limit":11}},"expected":{"Decide":{"approved":false,"reason":"limit"}}},
{"inputs":{"Decide":{"quantity":7,"limit":7}},"expected":{"Decide":{"approved":true,"reason":"available"}}},
{"inputs":{"Decide":{"quantity":5,"limit":8}},"expected":{"Decide":{"approved":true,"reason":"available"}}}
]}`)
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "checks973.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	original := compositionDigest(prior)
	run, err := ExecuteComposition(ctx, "checks973.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 4 || run.FiniteTotal != 4 || !run.RuntimeReplayed || len(run.ConditionSites) != 2 ||
		run.ModelCalls != 0 || run.InputSeparation.OverlappingInputs != 0 || run.InputSeparation.DisjointInputs != 4 {
		t.Fatal("selected record reason or actual condition evidence lost", err, run.Stage)
	}
	want := [][]bool{{true}, {false, true}, {false, false}, {false, false}}
	for c, values := range want {
		trace := run.Traces[c]
		if !trace.ConditionsObserved || len(trace.Conditions) != len(values) || len(trace.Deliveries[0].ActualFields) != 2 {
			t.Fatal("native field or condition denominator differs", trace)
		}
		for i, value := range values {
			event := trace.Conditions[i]
			if event.Value != value || event.SiteIndex != i {
				t.Fatalf("case %d actual branch %d differs: %+v", c, i, event)
			}
		}
	}
	for _, site := range run.ConditionSites {
		if site.ProjectionSHA256 != prior.GeneratedSHA256 || site.ActivityID != "checks973://decide" ||
			prior.Source[site.Start:site.End] != site.Expression {
			t.Fatal("record condition site lacks selected pure source evidence", site)
		}
	}
	if !strings.Contains(prior.GoooSource, "available") || compositionDigest(prior) != original ||
		VerifyComposition(ctx, "checks973.gooo", source, prior) != nil {
		t.Fatal("native observation changed the selected source")
	}
	saveConditionObservation(t, "checks973", source, run)
	t.Log("checks973: native records4/4 fields8/8, evaluated conditions7, declared sites2, observed outcomes4/4; disjoint inputs4; source unchanged; model0")
}

func saveConditionObservation(t *testing.T, name string, source []byte, run CompositionRuntime) {
	t.Helper()
	if root := os.Getenv("GOOO_CONDITION_OBSERVATION_OUTPUT"); root != "" {
		for suffix, value := range map[string]any{"runtime.json": run, "source.json": string(source)} {
			raw, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			file, err := os.OpenFile(root+"/"+name+"-"+suffix, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				t.Fatal(err)
			}
			_, writeErr := file.Write(append(raw, '\n'))
			closeErr := file.Close()
			if writeErr != nil || closeErr != nil {
				t.Fatal(writeErr, closeErr)
			}
		}
	}
}
