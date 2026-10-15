package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNativeIfConditionEvidenceProbe907(t *testing.T) {
	source := []byte(`package permit907
namespace permit907
entity Integer id "permit907://integer"
activity Decide(Integer) -> Integer id "permit907://decide" computes "if input < 17 { return 120 }; return input + 1"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Decide":12},"expected":{"Decide":120}},
{"inputs":{"Decide":21},"expected":{"Decide":22}}
]}`)
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "permit907.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(ctx, "permit907.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 2 || !run.RuntimeReplayed || run.ModelCalls != 0 {
		t.Fatal("native outputs unavailable", err)
	}
	raw, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Sites  []json.RawMessage `json:"condition_sites"`
		Traces []struct {
			Observed   bool `json:"conditions_observed"`
			Conditions []struct {
				Value bool `json:"value"`
			} `json:"conditions"`
		} `json:"traces"`
	}
	if err := json.Unmarshal(raw, &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.Sites) != 1 || len(evidence.Traces) != 2 ||
		!evidence.Traces[0].Observed || len(evidence.Traces[0].Conditions) != 1 ||
		!evidence.Traces[0].Conditions[0].Value || !evidence.Traces[1].Observed ||
		len(evidence.Traces[1].Conditions) != 1 || evidence.Traces[1].Conditions[0].Value {
		t.Fatalf("NATIVE_IF_TRACE_MISSING: native outputs 2/2, condition sites=%d; observed=%+v", len(evidence.Sites), evidence.Traces)
	}
}
