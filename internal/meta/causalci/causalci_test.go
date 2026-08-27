package causalci

import (
	"encoding/json"
	"testing"
)

func TestRawObservationHasNoConclusionSurface(t *testing.T) {
	statusDigest, err := digestJSON([]string{})
	if err != nil {
		t.Fatal(err)
	}
	valid := Observation{
		Schema: ObservationSchema, Repository: "example/repository", BaseSHA: "base", HeadSHA: "head", SourcePath: "policy.gooo",
		ChangedFiles: []ChangedFileObservation{{Path: "policy.gooo", Status: "M"}},
		PriorClaims:  []PriorClaimObservation{{ClaimID: "claim:causal-selection", SubjectPath: "policy.gooo", State: ClaimOpen, Provenance: "git://observation/policy.gooo"}},
		Isolation:    IsolationObservation{Before: RepositorySnapshot{StatusLines: []string{}, StatusDigest: statusDigest}, After: RepositorySnapshot{StatusLines: []string{}, StatusDigest: statusDigest}},
	}
	raw, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeObservation(raw); err != nil {
		t.Fatal(err)
	}
	var conclusion map[string]any
	if err := json.Unmarshal(raw, &conclusion); err != nil {
		t.Fatal(err)
	}
	conclusion["decision"] = "SELECTED"
	concluded, err := json.Marshal(conclusion)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeObservation(concluded); err == nil {
		t.Fatal("raw observation accepted a conclusion field")
	}
}
