package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConditionOriginMetaprogramBuiltinText2513(t *testing.T) {
	source := []byte(`package origin2513
namespace origin2513
entity Request id "origin2513://request" fields {
 field value id "origin2513://value" type "urn:gooo:type:integer" required one
 field enabled id "origin2513://enabled" type "urn:gooo:type:boolean" required one
}
entity Integer id "origin2513://integer"
activity Calculate(Request) -> Integer id "origin2513://calculate" computes "if input.enabled && input.value < 61 { return input.value + 12 }; return input.value - 12"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Calculate":{"value":12,"enabled":true}},"expected":{"Calculate":24}}]}`)
	ctx := context.Background()
	producer, err := GenerateComposition(ctx, "origin2513.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	sites, err := compositionConditionSites(producer)
	if err != nil {
		t.Fatal(err)
	}
	// Compiler metadata supplies this fact before executing the producer body.
	facts := ConditionOriginFacts(CompositionRuntime{ConditionSites: sites})
	if len(facts) != 1 || facts[0].Fields != 2 || facts[0].Located != 2 || !facts[0].Selected || !facts[0].Declaration {
		t.Fatal("static source origin differs", facts)
	}
	path := "../../examples/condition-origins/classify.gooo.fixture"
	if pending := os.Getenv("GOOO_ORIGIN_EXAMPLE_SOURCE"); pending != "" {
		path = pending
	}
	consumer, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	inputs := append(facts,
		ConditionOriginFact{SiteID: "fixture:original2513", Selected: true, Original: true, Declaration: true, Fields: 3, Located: 3},
		ConditionOriginFact{SiteID: "fixture:missing2513"},
		ConditionOriginFact{SiteID: "fixture:field2513", Selected: true, Declaration: true, Fields: 3, Located: 2},
		ConditionOriginFact{SiteID: "fixture:declaration2513", Selected: true})
	expected := []string{"selected-source", "source-retained", "origin-unresolved", "origin-unresolved", "origin-unresolved"}
	cases := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for i, fact := range inputs {
		raw, _ := json.Marshal(fact)
		want, _ := json.Marshal(expected[i])
		one := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Describe":`+string(raw)+`},"expected":{"Describe":`+string(want)+`}}]}`)
		cases.Cases = append(cases.Cases, one.Cases[0])
	}
	prior, run := newSourceAnchorRun(t, "originconsumer2513", consumer, cases)
	saveConditionObservation(t, "originconsumer2513", consumer, run)
	if run.FinitePassed != 5 || run.ModelCalls != 0 {
		t.Fatal("Gooo source did not interpret origins", run.FinitePassed)
	}
	if output := os.Getenv("GOOO_CONDITION_OBSERVATION_OUTPUT"); output != "" {
		data, _ := json.MarshalIndent(map[string]any{"actual_source_facts": facts, "source_fact_inputs": 1, "interpretation_fixtures": 4,
			"producer_body_executed": false, "consumer_runtime": run, "producer_source_sha256": producer.SelectedSourceSHA256,
			"consumer_source_sha256": prior.SelectedSourceSHA256}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "origin-usage2513.json"), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("origin2513: static source fact1 with2 field declaration origins; Gooo interprets1 source fact and4 explicit fixtures; native consumer5/5; model0")
}
