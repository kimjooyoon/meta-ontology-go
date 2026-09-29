package lsp

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func executionOriginReceiptFixturePart01(runtimePlanDigest string) valueexecution.ExecutionOriginReceipt {
	origin := provenance.ObserveOriginChain(provenance.OriginChain{
		DeclarationURI:        "file:///fixture.gooo",
		DeclarationSymbol:     "activity Execute",
		IRNode:                "activity:Execute",
		GeneratedURI:          "runtime-plan.json",
		GeneratedSymbol:       "runtime-plan:Execute",
		ReverseObservationURI: "execution-receipt.json",
		ReverseObservation:    "execution.completed",
		MetricName:            "gooo.execution-plan-replay.v1",
		MetricValue:           "COMPLETE",
		EvidenceDigest:        "fixture-evidence",
	})
	return valueexecution.ObserveExecutionOrigin(origin, valueexecution.Execution{
		ExecutionDigest:   "sha256:" + strings.Repeat("b", 64),
		RuntimePlanDigest: runtimePlanDigest,
		Phase:             valueexecution.ExecutionPhaseCompleted,
	})
}

func completeExecutionOriginReceiptFixturePart01() valueexecution.ExecutionOriginReceipt {
	return executionOriginReceiptFixturePart01("sha256:" + strings.Repeat("a", 64))
}

func completeGeneratedReplayEvidenceFixturePart01() GeneratedReplayEvidencePart01 {
	return GeneratedReplayEvidencePart01{
		SourceDigest:             "sha256:" + strings.Repeat("c", 64),
		SemanticDigest:           "sha256:" + strings.Repeat("d", 64),
		TypedPlanDigest:          "sha256:" + strings.Repeat("e", 64),
		RuntimePlanDigest:        "sha256:" + strings.Repeat("a", 64),
		GeneratedArtifactDigest:  "sha256:" + strings.Repeat("f", 64),
		ReverseObservationDigest: "sha256:" + strings.Repeat("1", 64),
	}
}

func TestObserveGeneratedReplayEvidenceReceiptClosurePart01BindsCompletedRuntimePlan(t *testing.T) {
	evidence := completeGeneratedReplayEvidenceFixturePart01()
	receipt := completeExecutionOriginReceiptFixturePart01()

	observation := ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidence, receipt)

	if observation.Status != ExecutionEvidenceReceiptClosureComplete || observation.MissingStageIndex != -1 {
		t.Fatalf("closure=%+v, want COMPLETE with no missing stage", observation)
	}
	if !observation.NonAuthorizing ||
		!ValidateGeneratedReplayEvidenceReceiptClosurePart01(observation, evidence, receipt) {
		t.Fatalf("closure did not validate as non-authorizing evidence: %+v", observation)
	}
}

func TestGeneratedReplayEvidenceReceiptClosurePreservesFirstMissingStage(t *testing.T) {
	evidence := completeGeneratedReplayEvidenceFixturePart01()
	evidence.SemanticDigest = ""
	receipt := completeExecutionOriginReceiptFixturePart01()

	observation := ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidence, receipt)
	if observation.Status != ExecutionEvidenceReceiptClosureUnknown ||
		observation.MissingStageIndex != 1 ||
		observation.Reason != "GENERATED_REPLAY_SEMANTIC_DIGEST_MISSING" {
		t.Fatalf("closure=%+v, want first missing semantic stage", observation)
	}
	if !ValidateGeneratedReplayEvidenceReceiptClosurePart01(observation, evidence, receipt) {
		t.Fatal("unknown first-stage closure did not validate")
	}
}

func TestGeneratedReplayEvidenceReceiptClosureRejectsPlanMismatchAndTampering(t *testing.T) {
	evidence := completeGeneratedReplayEvidenceFixturePart01()
	receipt := executionOriginReceiptFixturePart01("sha256:" + strings.Repeat("2", 64))
	observation := ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidence, receipt)
	if observation.Status != ExecutionEvidenceReceiptClosureUnknown ||
		observation.MissingStageIndex != 3 || observation.Reason != "EXECUTION_RUNTIME_PLAN_DIGEST_MISMATCH" {
		t.Fatalf("closure=%+v, want runtime-plan mismatch at stage 3", observation)
	}
	if !ValidateGeneratedReplayEvidenceReceiptClosurePart01(observation, evidence, receipt) {
		t.Fatal("runtime-plan mismatch closure did not validate")
	}

	receipt = completeExecutionOriginReceiptFixturePart01()
	receipt.ExecutionDigest = "tampered"
	observation = ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidence, receipt)
	if observation.Status != ExecutionEvidenceReceiptClosureUnknown ||
		observation.Reason != "EXECUTION_ORIGIN_RECEIPT_INVALID" {
		t.Fatalf("closure=%+v, want invalid-receipt UNKNOWN", observation)
	}
}

func TestGeneratedReplayDigestsRequireCurrentSourceAndPlanIdentities(t *testing.T) {
	evidence := completeGeneratedReplayEvidenceFixturePart01()
	receipt := completeExecutionOriginReceiptFixturePart01()
	params := ExecutionPlanProvenanceParamsPart01{
		GeneratedReplayEvidence: &evidence,
		ExecutionOriginReceipt:  &receipt,
	}
	generated, reverse, closure := executionPlanGeneratedReplayDigestsPart01(
		params,
		evidence.SourceDigest,
		evidence.SemanticDigest,
		evidence.TypedPlanDigest,
	)
	if generated == "" || reverse == "" || closure.Status != ExecutionEvidenceReceiptClosureComplete {
		t.Fatalf("valid replay digests = (%q, %q), closure=%+v", generated, reverse, closure)
	}

	generated, reverse, closure = executionPlanGeneratedReplayDigestsPart01(
		params,
		"sha256:"+strings.Repeat("9", 64),
		evidence.SemanticDigest,
		evidence.TypedPlanDigest,
	)
	if generated != "" || reverse != "" || closure.MissingStageIndex != 0 ||
		closure.Reason != "GENERATED_REPLAY_SOURCE_DIGEST_MISMATCH" {
		t.Fatalf("stale source was not kept UNKNOWN: generated=%q reverse=%q closure=%+v", generated, reverse, closure)
	}

	evidence.ReverseObservationDigest = ""
	params.GeneratedReplayEvidence = &evidence
	generated, reverse, closure = executionPlanGeneratedReplayDigestsPart01(
		params,
		evidence.SourceDigest,
		evidence.SemanticDigest,
		evidence.TypedPlanDigest,
	)
	if generated == "" || reverse != "" || closure.MissingStageIndex != 5 {
		t.Fatalf(
			"missing reverse observation lost its frontier: generated=%q reverse=%q closure=%+v",
			generated,
			reverse,
			closure,
		)
	}
}
