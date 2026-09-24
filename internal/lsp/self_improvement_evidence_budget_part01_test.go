package lsp

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestObserveSelfImprovementEvidenceBudgetLSP(t *testing.T) {
	source := valueexecution.ObserveSelfImprovementEvidenceBudget(
		valueexecution.SelfImprovementEvidenceBudgetInput{
			ForwardEvidenceCount:  2,
			ReverseEvidenceCount:  1,
			RequiredEvidenceCount: 3,
			ForwardDigest:         "forward",
			ReverseDigest:         "reverse",
		},
	)

	visible := ObserveSelfImprovementEvidenceBudgetLSP(
		"file:///workspace/main.gooo",
		7,
		"Improve",
		source,
	)
	if !visible.Visible {
		t.Fatal("adequate source observation should be visible")
	}
	if visible.Status != valueexecution.SelfImprovementEvidenceBudgetAdequate {
		t.Fatalf("status = %q, want %q", visible.Status, valueexecution.SelfImprovementEvidenceBudgetAdequate)
	}
	if visible.EvidenceCount != 3 {
		t.Fatalf("evidence count = %d, want 3", visible.EvidenceCount)
	}
	if !visible.NonAuthorizing {
		t.Fatal("LSP observation must remain non-authorizing")
	}
	if len(visible.Digest) != 64 || strings.Trim(visible.Digest, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want lowercase sha256", visible.Digest)
	}

	unknown := ObserveSelfImprovementEvidenceBudgetLSP("", 7, "Improve", source)
	if unknown.Visible {
		t.Fatal("invalid LSP context must not be visible")
	}
	if unknown.Status != valueexecution.SelfImprovementEvidenceBudgetUnknown {
		t.Fatalf("status = %q, want UNKNOWN", unknown.Status)
	}
}