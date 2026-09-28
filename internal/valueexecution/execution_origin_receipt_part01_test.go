package valueexecution

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestObserveExecutionOriginBindsPhaseWithoutGrantingAuthority(t *testing.T) {
	origin := provenance.ObserveOriginChain(provenance.OriginChain{
		DeclarationURI:        "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol:     "activity TypedChain",
		IRNode:                "activity:TypedChain",
		GeneratedURI:          "cmd/gooo/run_source_typed_chain_part01_test.go",
		GeneratedSymbol:       "TestTypedChain",
		ReverseObservationURI: "internal/valueexecution/replay.go",
		ReverseObservation:    "execution.replay.completed",
		MetricName:            "gooo.provenance.origin-chain.v1",
		MetricValue:           "complete=1",
		EvidenceDigest:        "sha256:typed-chain-observation",
	})
	execution := Execution{
		ExecutionDigest:   "sha256:execution-observation",
		RuntimePlanDigest: "sha256:runtime-plan-observation",
		Phase:             ExecutionPhaseCompleted,
		Conditions: []ExecutionCondition{
			{Type: ExecutionConditionComplete, Status: ExecutionConditionTrue, Reason: "EXECUTION_FINISHED"},
			{Type: ExecutionConditionPlanReady, Status: ExecutionConditionTrue, Reason: "COMPILED_PLAN_VALIDATED"},
		},
	}

	receipt := ObserveExecutionOrigin(origin, execution)
	if receipt.Status != ExecutionOriginStatusBound || receipt.Reason != "ORIGIN_BOUND_TO_EXECUTION" {
		t.Fatalf("receipt=%#v", receipt)
	}
	if receipt.Phase != ExecutionPhaseCompleted || receipt.ReceiptDigest == "" || receipt.RuntimePlanDigest != execution.RuntimePlanDigest {
		t.Fatalf("receipt lost execution identity: %#v", receipt)
	}
	if !receipt.NonAuthorizing || len(receipt.Conditions) != 2 {
		t.Fatalf("receipt guardrails/conditions=%#v", receipt)
	}

	execution.Phase = ExecutionPhaseFailed
	execution.ExecutionDigest = "sha256:failed-execution-observation"
	failedReceipt := ObserveExecutionOrigin(origin, execution)
	if failedReceipt.Status != ExecutionOriginStatusBound || failedReceipt.Phase != ExecutionPhaseFailed {
		t.Fatalf("failed phase was not preserved: %#v", failedReceipt)
	}
}

func TestObserveExecutionOriginKeepsIncompleteEvidenceUnknown(t *testing.T) {
	partial := provenance.ObserveOriginChain(provenance.OriginChain{
		DeclarationURI:    "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol: "activity TypedChain",
	})
	receipt := ObserveExecutionOrigin(partial, Execution{
		ExecutionDigest: "sha256:execution-observation",
		Phase:           ExecutionPhaseCompleted,
	})
	if receipt.Status != ExecutionOriginStatusUnknown || receipt.Reason != "ORIGIN_OBSERVATION_INCOMPLETE" {
		t.Fatalf("partial origin receipt=%#v", receipt)
	}
	empty := ObserveExecutionOrigin(provenance.OriginChainObservation{}, Execution{Phase: ExecutionPhaseCompleted})
	if empty.Status != ExecutionOriginStatusUnknown || empty.Reason != "ORIGIN_OBSERVATION_INCOMPLETE" {
		t.Fatalf("empty origin receipt=%#v", empty)
	}
}
