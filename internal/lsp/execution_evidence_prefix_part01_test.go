package lsp

import "testing"

func TestObserveExecutionEvidencePrefixComplete(t *testing.T) {
	observation := ObserveExecutionEvidencePrefix([]string{"completion", "receipt", "execution"})

	if observation.Status != ExecutionEvidencePrefixComplete {
		t.Fatalf("status = %q, want COMPLETE", observation.Status)
	}
	if observation.MissingStageIndex != -1 {
		t.Fatalf("missing stage index = %d, want -1", observation.MissingStageIndex)
	}
	if !observation.NonAuthorizing {
		t.Fatal("complete observation must remain non-authorizing")
	}
	if !ValidateExecutionEvidencePrefixObservation(observation) {
		t.Fatal("complete observation should validate")
	}
}

func TestObserveExecutionEvidencePrefixPreservesFirstMissingStage(t *testing.T) {
	observation := ObserveExecutionEvidencePrefix([]string{"completion", "", "execution"})

	if observation.Status != ExecutionEvidencePrefixUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
	if observation.MissingStageIndex != 1 {
		t.Fatalf("missing stage index = %d, want 1", observation.MissingStageIndex)
	}
	if observation.EvidencePrefixDigest != ObserveExecutionEvidencePrefix([]string{"completion"}).EvidencePrefixDigest {
		t.Fatal("prefix digest must stop at the first missing stage")
	}
	if !ValidateExecutionEvidencePrefixObservation(observation) {
		t.Fatal("missing-stage observation should validate")
	}
}

func TestValidateExecutionEvidencePrefixObservationRejectsTampering(t *testing.T) {
	observation := ObserveExecutionEvidencePrefix([]string{"completion", "receipt", "execution"})
	observation.EvidencePrefixDigest = "tampered"

	if ValidateExecutionEvidencePrefixObservation(observation) {
		t.Fatal("tampered prefix evidence must be rejected")
	}
}
