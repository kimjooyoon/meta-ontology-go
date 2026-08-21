package metarecognition

import (
	"reflect"
	"testing"
)

func TestGoooIgnoresExpectedMutation(t *testing.T) {
	caseValue := Corpus()[0]
	before, _ := Evaluate(caseValue)
	caseValue.Expected.Reason = ReasonExternalMissing
	caseValue.Expected.LocalizedIDs = []string{"forged://id"}
	after, _ := Evaluate(caseValue)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("gooo changed after expected mutation: before=%#v after=%#v", before, after)
	}
	manifest, err := Run([]Case{caseValue})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Summary.ExactReasonLocalizationVector {
		t.Fatal("manifest accepted forged expected reason/localization")
	}
}
func TestBaselineCountersDoNotAffectGooo(t *testing.T) {
	caseValue := Corpus()[0]
	beforeGooo, beforeBaseline := Evaluate(caseValue)
	caseValue.Baseline.FullCommands = 99
	caseValue.Baseline.SelectedCommands = 88
	caseValue.Baseline.ProvRecords = 77
	caseValue.Baseline.ProvPaths = 66
	afterGooo, afterBaseline := Evaluate(caseValue)
	if !reflect.DeepEqual(beforeGooo, afterGooo) {
		t.Fatalf("gooo copied baseline counters: before=%#v after=%#v", beforeGooo, afterGooo)
	}
	if reflect.DeepEqual(beforeBaseline, afterBaseline) {
		t.Fatal("baseline counters did not affect baseline accounting")
	}
}
func TestExternalCompleteness(t *testing.T) {
	values := []struct {
		name     string
		external ExternalAssertion
		id       string
	}{
		{name: "all-zero", external: ExternalAssertion{}, id: "external-authenticity"},
		{name: "authenticity", external: ExternalAssertion{Provider: true, Phase: true, Observer: true}, id: "external-authenticity"},
		{name: "provider", external: ExternalAssertion{Authenticity: true, Phase: true, Observer: true}, id: "external-provider"},
		{name: "phase", external: ExternalAssertion{Authenticity: true, Provider: true, Observer: true}, id: "external-phase"},
		{name: "observer", external: ExternalAssertion{Authenticity: true, Provider: true, Phase: true}, id: "external-observer"},
	}
	for _, value := range values {
		t.Run(value.name, func(t *testing.T) {
			caseValue := Corpus()[7]
			caseValue.Baseline.External = value.external
			gooo, baseline := Evaluate(caseValue)
			for name, outcome := range map[string]Outcome{"gooo": gooo, "baseline": baseline} {
				if outcome.State != UnknownFullSuiteRequired || outcome.Reason != ReasonExternalMissing || !equalIDs(outcome.LocalizedIDs, []string{value.id}) {
					t.Errorf("%s = %#v, want UNKNOWN external %s", name, outcome, value.id)
				}
			}
		})
	}
}
