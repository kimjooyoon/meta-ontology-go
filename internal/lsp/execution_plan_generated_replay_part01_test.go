package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func generatedReplayEvidencePart02() GeneratedReplayEvidencePart01 {
	return GeneratedReplayEvidencePart01{
		SourceDigest:             "sha256:source",
		SemanticDigest:           "sha256:semantic",
		TypedPlanDigest:          "sha256:typed-plan",
		RuntimePlanDigest:        "sha256:runtime-plan",
		GeneratedArtifactDigest:  "sha256:generated",
		ReverseObservationDigest: "sha256:reverse",
	}
}

func generatedReplayReceiptPart02() valueexecution.ExecutionOriginReceipt {
	return valueexecution.ExecutionOriginReceipt{
		ExecutionDigest:   "sha256:execution",
		RuntimePlanDigest: "sha256:runtime-plan",
		Phase:             valueexecution.ExecutionPhaseCompleted,
		Status:            valueexecution.ExecutionOriginStatusBound,
		NonAuthorizing:    true,
		ReceiptDigest:     "sha256:receipt",
	}
}

func TestExecutionPlanGeneratedReplayDigestsRequireClosedReceipt(t *testing.T) {
	evidence := generatedReplayEvidencePart02()
	receipt := generatedReplayReceiptPart02()
	params := ExecutionPlanProvenanceParamsPart01{
		GeneratedReplayEvidence: &evidence,
		ExecutionOriginReceipt:  &receipt,
	}
	generated, reverse := executionPlanGeneratedReplayDigestsPart01(params)
	if generated != evidence.GeneratedArtifactDigest || reverse != evidence.ReverseObservationDigest {
		t.Fatalf("closed receipt digests=(%q, %q), want (%q, %q)", generated, reverse, evidence.GeneratedArtifactDigest, evidence.ReverseObservationDigest)
	}

	evidence.ReverseObservationDigest = ""
	generated, reverse = executionPlanGeneratedReplayDigestsPart01(ExecutionPlanProvenanceParamsPart01{
		GeneratedReplayEvidence: &evidence,
		ExecutionOriginReceipt:  &receipt,
	})
	if generated != "" || reverse != "" {
		t.Fatalf("incomplete replay digests=(%q, %q), want empty UNKNOWN boundary", generated, reverse)
	}
}
