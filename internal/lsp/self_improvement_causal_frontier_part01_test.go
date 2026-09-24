package lsp

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestObserveSelfImprovementCausalFrontierLSP(t *testing.T) {
	source := valueexecution.ObserveSelfImprovementCausalFrontier(
		[]valueexecution.SelfImprovementCausalFrontierCandidate{
			{
				ID:              "refuted",
				State:           "REFUTED",
				EvidenceCount:   2,
				DependencyCount: 1,
				Digest:          "refuted-digest",
			},
		},
	)
	observation := ObserveSelfImprovementCausalFrontierLSP(
		"file:///workspace/main.gooo",
		9,
		"Improve",
		source,
	)
	if !observation.Visible {
		t.Fatal("selected frontier should be visible")
	}
	if observation.SelectedID != "refuted" {
		t.Fatalf("selected id = %q, want refuted", observation.SelectedID)
	}
	if observation.Status != valueexecution.SelfImprovementCausalFrontierSelected {
		t.Fatalf("status = %q, want SELECTED", observation.Status)
	}
	if !observation.NonAuthorizing {
		t.Fatal("LSP observation must remain non-authorizing")
	}
	if len(observation.Digest) != 64 || strings.Trim(observation.Digest, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want lowercase sha256", observation.Digest)
	}
}

func TestObserveSelfImprovementCausalFrontierLSPRejectsInvalidContext(t *testing.T) {
	source := valueexecution.ObserveSelfImprovementCausalFrontier(nil)
	observation := ObserveSelfImprovementCausalFrontierLSP("", 9, "Improve", source)
	if observation.Visible {
		t.Fatal("invalid LSP context must not be visible")
	}
	if observation.Status != valueexecution.SelfImprovementCausalFrontierUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
}