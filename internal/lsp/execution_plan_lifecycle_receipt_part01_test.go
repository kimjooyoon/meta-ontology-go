package lsp

import "testing"

func TestExecutionPlanLifecycleReceiptPart01CompleteBoundary(t *testing.T) {
	receipt := ExecutionPlanLifecycleReceiptPart01{
		Schema:                   executionPlanProvenanceBindingSchemaPart01,
		State:                    ExecutionPlanLifecycleCompletePart01,
		StageIndex:               5,
		MissingStageIndex:        -1,
		EvidencePrefixDigest:     "prefix",
		SourceDigest:             "source",
		SemanticDigest:           "semantic",
		TypedPlanDigest:          "typed-plan",
		RuntimePlanDigest:        "runtime-plan",
		GeneratedArtifactDigest:  "generated",
		ReverseObservationDigest: "reverse",
	}
	if !receipt.ValidPart01() || !receipt.CompletePart01() {
		t.Fatalf("complete lifecycle receipt should be valid: %+v", receipt)
	}
	if got := receipt.FirstMissingStagePart01(); got != -1 {
		t.Fatalf("missing stage = %d, want -1", got)
	}
}

func TestExecutionPlanLifecycleReceiptPart01UnknownPreservesFrontier(t *testing.T) {
	receipt := ExecutionPlanLifecycleReceiptPart01{
		Schema:               executionPlanProvenanceBindingSchemaPart01,
		State:                ExecutionPlanLifecycleUnknownPart01,
		StageIndex:           4,
		MissingStageIndex:    4,
		EvidencePrefixDigest: "prefix-4",
		SourceDigest:         "source",
		SemanticDigest:       "semantic",
	}
	if !receipt.ValidPart01() {
		t.Fatalf("unknown lifecycle receipt should preserve a valid frontier: %+v", receipt)
	}
	if receipt.CompletePart01() {
		t.Fatal("unknown lifecycle receipt must not close as complete")
	}
	if got := receipt.FirstMissingStagePart01(); got != 4 {
		t.Fatalf("missing stage = %d, want 4", got)
	}
}

func TestExecutionPlanLifecycleReceiptPart01RejectsTamperedPrefix(t *testing.T) {
	receipt := ExecutionPlanLifecycleReceiptPart01{
		Schema:               executionPlanProvenanceBindingSchemaPart01,
		State:                ExecutionPlanLifecycleSuspendedPart01,
		StageIndex:           3,
		MissingStageIndex:    3,
		EvidencePrefixDigest: " ",
	}
	if receipt.ValidPart01() {
		t.Fatal("blank evidence prefix must not validate")
	}
}
