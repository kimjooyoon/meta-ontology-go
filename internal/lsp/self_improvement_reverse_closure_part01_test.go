package lsp

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestObserveSelfImprovementReverseClosureLSP(t *testing.T) {
	source := valueexecution.SelfImprovementReverseClosureObservation{
		Status:             valueexecution.SelfImprovementReverseClosureCompleted,
		Outcome:            "COMPLETED",
		ReverseObservation: true,
		Digest:             "reverse-digest",
	}
	observation := ObserveSelfImprovementReverseClosureLSP(
		"file:///workspace/main.gooo",
		11,
		"Improve",
		source,
	)
	if !observation.Visible {
		t.Fatal("completed reverse closure should be visible")
	}
	if !observation.ReverseObservation {
		t.Fatal("reverse observation must remain true")
	}
	if observation.Outcome != "COMPLETED" {
		t.Fatalf("outcome = %q, want COMPLETED", observation.Outcome)
	}
	if !observation.NonAuthorizing {
		t.Fatal("LSP observation must remain non-authorizing")
	}
	if len(observation.Digest) != 64 || strings.Trim(observation.Digest, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want lowercase sha256", observation.Digest)
	}
}

func TestObserveSelfImprovementReverseClosureLSPRejectsUnknownSource(t *testing.T) {
	observation := ObserveSelfImprovementReverseClosureLSP(
		"file:///workspace/main.gooo",
		11,
		"Improve",
		valueexecution.SelfImprovementReverseClosureObservation{
			Status: valueexecution.SelfImprovementReverseClosureUnknown,
		},
	)
	if observation.Visible {
		t.Fatal("unknown source should not be visible")
	}
	if observation.Status != valueexecution.SelfImprovementReverseClosureUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
}