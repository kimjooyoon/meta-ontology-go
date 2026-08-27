package causalci

import (
	"encoding/json"
	"testing"
)

func TestRawObservationHasNoConclusionSurface(t *testing.T) {
	snapshotDigest, err := digestJSON([]RepositoryEntry{})
	if err != nil {
		t.Fatal(err)
	}
	claimID := ClaimInstanceID("claim-template:causal-route", "policy.gooo", ReasonCompleteRoute)
	valid := Observation{
		Schema: ObservationSchema, Repository: "example/repository", BaseSHA: "base", HeadSHA: "head", ObservedCheckoutSHA: "head", SourcePath: "policy.gooo", HeadPathObjectID: GitBlobObjectID([]byte("source")), SourceBytesDigest: digestBytes([]byte("source")),
		ChangedFiles: []ChangedFileObservation{{Path: "policy.gooo", Status: "M"}},
		PriorClaims:  []PriorClaimObservation{{TemplateID: "claim-template:causal-route", InstanceID: claimID, SubjectPath: "policy.gooo", Proposition: ReasonCompleteRoute, State: ClaimOpen, Provenance: "git://observation/policy.gooo"}},
		Isolation:    IsolationObservation{Before: RepositorySnapshot{Entries: []RepositoryEntry{}, SnapshotDigest: snapshotDigest}, After: RepositorySnapshot{Entries: []RepositoryEntry{}, SnapshotDigest: snapshotDigest}},
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
