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
		Schema: ObservationSchema, Repository: "example/repository", BaseSHA: "base", HeadSHA: "head", ObservedCheckoutSHA: "head", SourcePath: "policy.gooo", ObjectFormat: "sha1", HeadPathObjectID: GitBlobObjectID([]byte("source")), SourceBytesDigest: digestBytes([]byte("source")),
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

func TestContradictionTargetsOneSubjectAndClaimInstance(t *testing.T) {
	sourcePath, unrelatedPath := "policy.gooo", "unrelated.go"
	route := ReasonCompleteRoute
	sourceClaim := PriorClaimObservation{TemplateID: "claim-template:causal-route", SubjectPath: sourcePath, Proposition: route, State: ClaimOpen, Provenance: "git://observation/policy.gooo"}
	sourceClaim.InstanceID = ClaimInstanceID(sourceClaim.TemplateID, sourceClaim.SubjectPath, sourceClaim.Proposition)
	unrelatedClaim := PriorClaimObservation{TemplateID: "claim-template:causal-route", SubjectPath: unrelatedPath, Proposition: route, State: ClaimOpen, Provenance: "git://observation/unrelated.go"}
	unrelatedClaim.InstanceID = ClaimInstanceID(unrelatedClaim.TemplateID, unrelatedClaim.SubjectPath, unrelatedClaim.Proposition)
	observation := Observation{SourcePath: sourcePath, ChangedFiles: []ChangedFileObservation{{Path: sourcePath, Status: "M"}, {Path: unrelatedPath, Status: "M"}}, PriorClaims: []PriorClaimObservation{sourceClaim, unrelatedClaim}}
	policy := PolicyGraph{
		Source: SourceEvidence{Path: sourcePath},
		Edges:  []PolicyEdge{{ID: "policy-edge:exact"}},
		Contradictions: []PolicyContradiction{{
			Stage: stageConformance, Step: stepValidatePolicy, Reason: ReasonContradictoryPolicy,
			SubjectPath: sourcePath, ClaimInstanceIDs: []string{sourceClaim.InstanceID}, Edges: []string{"policy-edge:exact"},
		}},
	}
	subjects := evaluateSubjects(observation, policy)
	byPath := map[string]SubjectResolution{}
	for _, subject := range subjects {
		byPath[subject.Path] = subject
	}
	if subject := byPath[unrelatedPath]; subject.Resolution != ResolutionUnknown || subject.Coordinate.Step != stepDescendFull {
		t.Fatalf("unrelated subject was not lowered to full suite: %#v", subjects)
	}
	if subject := byPath[sourcePath]; subject.Resolution != ResolutionFailClosed {
		t.Fatalf("target subject was not fail-closed: %#v", subjects)
	}
	transitions := appendClaimTransitions(observation, policy, subjects, "sha256:observation")
	for _, transition := range transitions {
		switch transition.SubjectPath {
		case sourcePath:
			if transition.After != ClaimRefuted || transition.Resolution != PlanNone {
				t.Fatalf("target claim was not structurally refuted: %#v", transition)
			}
		case unrelatedPath:
			if transition.After != ClaimOpen || transition.Resolution != PlanFull || transition.Reason != ReasonClaimLowered {
				t.Fatalf("unrelated same-proposition claim did not remain open at lower resolution: %#v", transition)
			}
		}
	}
}

func TestMissingContradictionEdgeDoesNotBindSubjectOrClaim(t *testing.T) {
	sourcePath := "policy.gooo"
	claim := PriorClaimObservation{TemplateID: "claim-template:causal-route", SubjectPath: sourcePath, Proposition: ReasonCompleteRoute, State: ClaimOpen, Provenance: "git://observation/policy.gooo"}
	claim.InstanceID = ClaimInstanceID(claim.TemplateID, claim.SubjectPath, claim.Proposition)
	observation := Observation{SourcePath: sourcePath, ChangedFiles: []ChangedFileObservation{{Path: sourcePath, Status: "M"}}, PriorClaims: []PriorClaimObservation{claim}}
	policy := PolicyGraph{Source: SourceEvidence{Path: sourcePath}, Contradictions: []PolicyContradiction{{Stage: stageConformance, Step: stepValidatePolicy, Reason: "REQUIRED_CAUSAL_POLICY_VALUE_MISSING", SubjectPath: sourcePath}}}
	attachContradictionTargets(&policy, observation)
	if len(policy.Contradictions) != 1 || len(policy.Contradictions[0].Edges) != 0 || len(policy.Contradictions[0].ClaimInstanceIDs) != 0 {
		t.Fatalf("missing policy edge became evidence: %#v", policy.Contradictions)
	}
	subject := evaluateSubjects(observation, policy)[0]
	if subject.Resolution == ResolutionFailClosed {
		t.Fatalf("missing policy edge fail-closed a subject: %#v", subject)
	}
	transition := appendClaimTransitions(observation, policy, []SubjectResolution{subject}, "sha256:observation")[0]
	if transition.After != ClaimOpen || transition.Reason != ReasonClaimLowered {
		t.Fatalf("missing policy edge refuted or discharged claim: %#v", transition)
	}
	conformance := conformanceFor(policy, PlanGate{Decision: PlanGatePass})
	if conformance.Decision != ConformanceFailClosed || len(conformance.RootContradictionInventory) != 1 || conformance.RootContradictionInventoryDigest == "" {
		t.Fatalf("root contradiction inventory was not preserved: %#v", conformance)
	}
}
