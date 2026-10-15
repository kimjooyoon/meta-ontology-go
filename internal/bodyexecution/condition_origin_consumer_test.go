package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFreshConditionOriginConsumer2459(t *testing.T) {
	source := []byte(`package origin2459
namespace origin2459
entity Request id "origin2459://request" fields {
 field quantity id "origin2459://quantity" type "urn:gooo:type:integer" required one
}
entity Integer id "origin2459://integer"
activity Calculate(Request) -> Integer id "origin2459://calculate" computes "if input.quantity < 0 { return -2 }; if input.quantity > 52 { return input.quantity + 10 }; return input.quantity * 3"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Calculate":{"quantity":-3}},"expected":{"Calculate":-2}},{"inputs":{"Calculate":{"quantity":7}},"expected":{"Calculate":21}},{"inputs":{"Calculate":{"quantity":58}},"expected":{"Calculate":68}}]}`)
	prior, run := newSourceAnchorRun(t, "origin2459", source, suite)
	saveConditionObservation(t, "origin2459", source, run)
	facts := ConditionOriginFacts(run)
	if run.FinitePassed != 3 || len(facts) != 2 {
		t.Fatal("fresh record producer differs")
	}
	consumerPath := "../../examples/condition-origins/classify.gooo.fixture"
	if pending := os.Getenv("GOOO_ORIGIN_EXAMPLE_SOURCE"); pending != "" {
		consumerPath = pending
	}
	consumer, err := os.ReadFile(consumerPath)
	if err != nil {
		t.Fatal(err)
	}
	cases := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	add := func(fact ConditionOriginFact, expected string) {
		t.Helper()
		raw, _ := json.Marshal(fact)
		want, _ := json.Marshal(expected)
		one := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Describe":`+string(raw)+`},"expected":{"Describe":`+string(want)+`}}]}`)
		cases.Cases = append(cases.Cases, one.Cases[0])
	}
	for _, fact := range facts {
		if !fact.Selected || !fact.Declaration || fact.Original || fact.Fields != 1 || fact.Located != 1 {
			t.Fatal("adapter changed origin meaning", fact)
		}
		add(fact, "selected-source")
	}
	// These three inputs are explicit fixtures for the source's interpretation,
	// distinct from the two facts just produced by native execution.
	add(ConditionOriginFact{SiteID: "fixture:original", Selected: true, Original: true, Declaration: true, Fields: 2, Located: 2}, "source-retained")
	add(ConditionOriginFact{SiteID: "fixture:missing"}, "origin-unresolved")
	add(ConditionOriginFact{SiteID: "fixture:field", Selected: true, Declaration: true, Fields: 2, Located: 1}, "origin-unresolved")
	classification, err := GenerateComposition(context.Background(), "classify.gooo", consumer, cases, "")
	if err != nil {
		t.Fatal(err)
	}
	classified, err := ExecuteComposition(context.Background(), "classify.gooo", consumer, classification, cases, nativeTool())
	if err != nil || classified.FinitePassed != 5 {
		t.Fatal("Gooo interpretation differs", err, classified.FinitePassed)
	}
	if output := os.Getenv("GOOO_CONDITION_OBSERVATION_OUTPUT"); output != "" {
		data, _ := json.MarshalIndent(map[string]any{"facts": facts, "actual_origin_inputs": 2, "interpretation_fixtures": 3,
			"runtime": classified, "producer_source_sha256": prior.SelectedSourceSHA256, "consumer_source_sha256": classification.SelectedSourceSHA256}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "origin-usage2459.json"), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("origin2459: native3/3; source/field declarations retained2/2; Gooo interprets2 actual facts and3 fixtures; no model")
}
