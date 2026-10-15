package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConditionOriginsCopiedHelperAndGoooUsage2117(t *testing.T) {
	source := []byte(`package helpers2117
namespace helpers2117
entity Integer id "helpers2117://integer"
entity Boolean id "helpers2117://boolean"
activity Gate(Integer) -> Boolean id "helpers2117://gate" computes "if input < 41 && input >= 0 { return true }; return false"
activity First(Integer) -> Integer id "helpers2117://first" computes "if input < 0 { return input + 2 }; if Gate(input * 2) { return input + 8 }; return input - 8"
activity Second(Integer) -> Integer id "helpers2117://second" computes "let floor = 5; if Gate(input + floor) { return input * 5 }; return input - 7"
bind First.result -> Second.input
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"First":8},"expected":{"First":16,"Second":80}},{"inputs":{"First":39},"expected":{"First":31,"Second":155}},{"inputs":{"First":-6},"expected":{"First":-4,"Second":-20}}]}`)
	prior, run := newSourceAnchorRun(t, "helpers2117", source, suite, "Second")
	saveConditionObservation(t, "helpers2117", source, run)
	if run.FinitePassed != 6 || len(run.ConditionSites) != 5 {
		t.Fatal("fresh entry-scoped helper behavior differs", run.FinitePassed, len(run.ConditionSites))
	}
	starts, roots := map[int]bool{}, map[string]bool{}
	events := 0
	for _, trace := range run.Traces {
		events += len(trace.Conditions)
	}
	for _, site := range run.ConditionSites {
		if site.Source == nil || site.Source.Status != "MAPPED" || site.Source.Declaration == nil || site.Source.Declaration.StableID != site.ActivityID || site.ExpressionID != conditionSiteIdentity(site) {
			t.Fatal("source declaration or expression identity differs", site)
		}
		if site.ActivityID == "helpers2117://gate" {
			starts[site.Source.Start], roots[site.RootActivityID] = true, true
		}
	}
	if len(starts) != 1 || len(roots) != 2 || events != 13 {
		t.Fatal("shared helper origin collapsed", starts, roots, events)
	}
	consumerPath := "../../examples/condition-origins/classify.gooo.fixture"
	if pending := os.Getenv("GOOO_ORIGIN_EXAMPLE_SOURCE"); pending != "" {
		consumerPath = pending
	}
	consumer, err := os.ReadFile(consumerPath)
	if err != nil {
		t.Fatal(err)
	}
	facts := ConditionOriginFacts(run)
	consumerCases := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, fact := range facts {
		raw, _ := json.Marshal(fact)
		testCase, err := DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Describe":` + string(raw) + `},"expected":{"Describe":"selected-source"}}]}`))
		if err != nil {
			t.Fatal(err)
		}
		consumerCases.Cases = append(consumerCases.Cases, testCase.Cases[0])
	}
	classified, err := GenerateComposition(context.Background(), "classify.gooo", consumer, consumerCases, "")
	if err != nil {
		t.Fatal(err)
	}
	classifiedRun, err := ExecuteComposition(context.Background(), "classify.gooo", consumer, classified, consumerCases, nativeTool())
	if err != nil || classifiedRun.FinitePassed != 5 {
		t.Fatal("Gooo source did not consume origin facts", err, classifiedRun.FinitePassed)
	}
	if output := os.Getenv("GOOO_CONDITION_OBSERVATION_OUTPUT"); output != "" {
		data, _ := json.MarshalIndent(map[string]any{"facts": facts, "runtime": classifiedRun, "producer_source_sha256": prior.SelectedSourceSHA256, "consumer_source_sha256": classified.SelectedSourceSHA256}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "origin-usage2117.json"), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("helpers2117: helper origin across2 roots; native6/6 events13; fresh origin facts classified by Gooo5/5; no model")
}
