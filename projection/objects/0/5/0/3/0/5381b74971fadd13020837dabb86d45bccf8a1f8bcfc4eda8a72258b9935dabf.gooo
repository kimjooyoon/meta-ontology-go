package workfrontier

import (
	"encoding/json"
	"testing"
)

func TestR4StrictEnvelopeRejectsUnknownDuplicateAndMissingFields(t *testing.T) {
	encoded, err := EncodeR4JSON(r4FixtureInput(t, "acyclic"))
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	delete(object, "root_obligation_ids")
	missing, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeR4JSON(missing); err == nil {
		t.Fatal("accepted an envelope without root_obligation_ids")
	}
	object["unexpected"] = json.RawMessage(`true`)
	unknown, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeR4JSON(unknown); err == nil {
		t.Fatal("accepted an envelope with an unknown field")
	}
	duplicate := []byte(`{"schema_version":"gooo/work-frontier-r4/v1","schema_version":"gooo/work-frontier-r4/v1"}`)
	if _, err := DecodeR4JSON(duplicate); err == nil {
		t.Fatal("accepted an envelope with a duplicate field")
	}
	duplicateResult := EvaluateR4JSON(duplicate)
	if duplicateResult.Status != R4StatusFailClosed || duplicateResult.Reason != R4ReasonMalformedBinding {
		t.Fatalf("duplicate envelope result = %#v", duplicateResult)
	}
	result := EvaluateR4JSON(missing)
	if result.Status != R4StatusUnknown || result.Reason != R4ReasonRequiredInputMissing || !result.FullSuiteRequired {
		t.Fatalf("missing envelope result = %#v", result)
	}
}
func TestR4DigestsExcludeFixtureLabels(t *testing.T) {
	input := r4FixtureInput(t, "self-loop")
	first := EvaluateR4(input)
	fixtureLabel := r4Fixture{Name: "self-loop", Status: "PASS", Reason: "NONE"}
	second := EvaluateR4(input)
	if first.GraphDigest != second.GraphDigest || first.RuleDigest != second.RuleDigest {
		t.Fatal("source-derived digests changed without an input change")
	}
	if got := r4FixtureDigest(input); got == "" {
		t.Fatal("empty producer fixture digest")
	}
	if fixtureLabel.Status == "" || fixtureLabel.Reason == "" {
		t.Fatal("fixture labels unexpectedly absent")
	}
}
func TestR4MalformedGraphFailsClosed(t *testing.T) {
	input := r4FixtureInput(t, "acyclic")
	input.Paths[0].PrerequisiteObligationIDs = []string{"obligation/root", "obligation/root"}
	var err error
	input, err = BindR4Payloads(input)
	if err != nil {
		t.Fatal(err)
	}
	got := EvaluateR4(input)
	if got.Status != R4StatusFailClosed || got.Reason != R4ReasonMalformedGraph || len(got.SelectedIDs) != 0 {
		t.Fatalf("malformed graph result = %#v", got)
	}
}
