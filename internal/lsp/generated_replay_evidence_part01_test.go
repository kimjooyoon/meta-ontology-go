package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func generatedReplayReceiptPart01() valueexecution.ExecutionOriginReceipt {
	return valueexecution.ExecutionOriginReceipt{
		ExecutionDigest:   "sha256:execution",
		RuntimePlanDigest: "sha256:runtime-plan",
		Phase:             valueexecution.ExecutionPhaseCompleted,
		Status:            valueexecution.ExecutionOriginStatusBound,
		NonAuthorizing:    true,
		ReceiptDigest:     "sha256:receipt",
	}
}

func completeGeneratedReplayEvidencePart01() GeneratedReplayEvidencePart01 {
	return GeneratedReplayEvidencePart01{
		SourceDigest:             "sha256:source",
		SemanticDigest:           "sha256:semantic",
		TypedPlanDigest:          "sha256:typed-plan",
		RuntimePlanDigest:        "sha256:runtime-plan",
		GeneratedArtifactDigest:  "sha256:generated",
		ReverseObservationDigest: "sha256:reverse",
	}
}

func TestObserveGeneratedReplayEvidenceReceiptClosurePart01Completes(t *testing.T) {
	evidence := completeGeneratedReplayEvidencePart01()
	receipt := generatedReplayReceiptPart01()

	observation := ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidence, receipt)

	if observation.Status != ExecutionEvidenceReceiptClosureComplete {
		t.Fatalf("status=%q, want COMPLETE", observation.Status)
	}
	if observation.MissingStageIndex != -1 || !observation.NonAuthorizing {
		t.Fatalf("closure boundary=%+v", observation)
	}
	if !ValidateExecutionEvidenceReceiptClosureObservation(
		observation,
		ObserveGeneratedReplayEvidencePrefixPart01(evidence),
		receipt,
	) {
		t.Fatal("generated replay closure did not validate")
	}
}

func TestObserveGeneratedReplayEvidenceReceiptClosurePart01PreservesFirstMissingStage(t *testing.T) {
	evidence := completeGeneratedReplayEvidencePart01()
	evidence.RuntimePlanDigest = ""
	receipt := generatedReplayReceiptPart01()

	prefix := ObserveGeneratedReplayEvidencePrefixPart01(evidence)
	observation := ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidence, receipt)

	if prefix.Status != ExecutionEvidencePrefixUnknown || prefix.MissingStageIndex != 3 {
		t.Fatalf("prefix boundary=%+v", prefix)
	}
	if observation.Status != ExecutionEvidenceReceiptClosureUnknown ||
		observation.MissingStageIndex != 3 ||
		observation.Reason != "EXECUTION_EVIDENCE_PREFIX_INCOMPLETE" {
		t.Fatalf("closure boundary=%+v", observation)
	}
}

func TestGeneratedReplayEvidencePart01KeepsFixedStageOrder(t *testing.T) {
	evidence := completeGeneratedReplayEvidencePart01()

	want := []string{
		"sha256:source",
		"sha256:semantic",
		"sha256:typed-plan",
		"sha256:runtime-plan",
		"sha256:generated",
		"sha256:reverse",
	}
	got := evidence.StageDigests()
	if len(got) != len(want) {
		t.Fatalf("stage count=%d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("stage %d=%q, want %q", index, got[index])
		}
	}
}