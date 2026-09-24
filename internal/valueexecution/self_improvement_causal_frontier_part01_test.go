package valueexecution

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementCausalFrontier(t *testing.T) {
	observation := ObserveSelfImprovementCausalFrontier(
		[]SelfImprovementCausalFrontierCandidate{
			{
				ID:              "closed",
				State:           "CLOSED",
				EvidenceCount:   1,
				DependencyCount: 0,
				Digest:          "closed-digest",
			},
			{
				ID:              "unknown",
				State:           "UNKNOWN",
				EvidenceCount:   5,
				DependencyCount: 4,
				Digest:          "unknown-digest",
			},
			{
				ID:              "refuted",
				State:           "REFUTED",
				EvidenceCount:   3,
				DependencyCount: 2,
				Digest:          "refuted-digest",
			},
		},
	)

	if observation.Status != SelfImprovementCausalFrontierSelected {
		t.Fatalf("status = %q, want SELECTED", observation.Status)
	}
	if observation.SelectedID != "refuted" {
		t.Fatalf("selected id = %q, want refuted", observation.SelectedID)
	}
	if observation.SelectedState != "REFUTED" {
		t.Fatalf("selected state = %q, want REFUTED", observation.SelectedState)
	}
	if !observation.NonAuthorizing {
		t.Fatal("causal frontier must remain non-authorizing")
	}
	if len(observation.Digest) != 64 || strings.Trim(observation.Digest, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want lowercase sha256", observation.Digest)
	}
}

func TestObserveSelfImprovementCausalFrontierRejectsInvalidCandidates(t *testing.T) {
	observation := ObserveSelfImprovementCausalFrontier(
		[]SelfImprovementCausalFrontierCandidate{
			{ID: "bad-state", State: "READY", Digest: "digest"},
			{ID: "missing-digest", State: "UNKNOWN"},
		},
	)
	if observation.Status != SelfImprovementCausalFrontierUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
	if observation.Reason != "CAUSAL_FRONTIER_EMPTY" {
		t.Fatalf("reason = %q, want CAUSAL_FRONTIER_EMPTY", observation.Reason)
	}
}