package valueexecution

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementEvidenceBudget(t *testing.T) {
	tests := []struct {
		name   string
		input  SelfImprovementEvidenceBudgetInput
		status SelfImprovementEvidenceBudgetStatus
		reason string
	}{
		{
			name: "adequate",
			input: SelfImprovementEvidenceBudgetInput{
				ForwardEvidenceCount:  2,
				ReverseEvidenceCount:  2,
				RequiredEvidenceCount: 3,
				ForwardDigest:         "forward",
				ReverseDigest:         "reverse",
			},
			status: SelfImprovementEvidenceBudgetAdequate,
			reason: "EVIDENCE_BUDGET_ADEQUATE",
		},
		{
			name: "insufficient",
			input: SelfImprovementEvidenceBudgetInput{
				ForwardEvidenceCount:  1,
				ReverseEvidenceCount:  1,
				RequiredEvidenceCount: 3,
				ForwardDigest:         "forward",
				ReverseDigest:         "reverse",
			},
			status: SelfImprovementEvidenceBudgetInsufficient,
			reason: "EVIDENCE_BUDGET_INSUFFICIENT",
		},
		{
			name: "unknown when digest is missing",
			input: SelfImprovementEvidenceBudgetInput{
				ForwardEvidenceCount:  2,
				ReverseEvidenceCount:  2,
				RequiredEvidenceCount: 3,
				ForwardDigest:         "forward",
			},
			status: SelfImprovementEvidenceBudgetUnknown,
			reason: "EVIDENCE_DIGEST_MISSING",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observation := ObserveSelfImprovementEvidenceBudget(test.input)
			if observation.Status != test.status {
				t.Fatalf("status = %q, want %q", observation.Status, test.status)
			}
			if observation.Reason != test.reason {
				t.Fatalf("reason = %q, want %q", observation.Reason, test.reason)
			}
			if !observation.NonAuthorizing {
				t.Fatal("observation must remain non-authorizing")
			}
			if len(observation.Digest) != 64 || strings.Trim(observation.Digest, "0123456789abcdef") != "" {
				t.Fatalf("digest = %q, want lowercase sha256", observation.Digest)
			}
		})
	}
}