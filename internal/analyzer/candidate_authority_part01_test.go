package analyzer

import (
	"testing"
)

func TestSemanticAdapterUnmappedCandidateRemainsDeferred(t *testing.T) {
	analysis := Result{
		Registrations: billingRegistrations(),
		Delta: SemanticDelta{
			Added: []Fact{{
				Subject:  NewIdentity("billing", "billing://activity/pay-order"),
				Relation: RelationUses, Object: NewIdentity("billing", "billing://entity/order"),
				Span: testSpan(), Origin: OriginSignature,
			}},
			Candidates: []Candidate{{
				Subject:  NewIdentity("billing", "billing://activity/pay-order"),
				Relation: RelationInvokes, Options: []Identity{
					NewIdentity("fraud", "fraud://activity/check"),
				}, Span: testSpan(), Reason: "unmapped relation",
			}},
		},
	}
	adapted := adaptAnalysis(t, analysis, billingPolicy(t, RelationUses))
	if len(adapted.IR.Graph.DeterministicFacts()) != 1 || len(adapted.IR.Graph.Candidates()) != 0 {
		t.Fatalf("graph views = %d deterministic, %d candidates", len(adapted.IR.Graph.DeterministicFacts()), len(adapted.IR.Graph.Candidates()))
	}
	if len(adapted.DeferredCandidates) != 1 || len(adapted.NormalizedDelta.CandidateFacts) != 1 {
		t.Fatalf("deferred candidate views = %d local, %d normalized", len(adapted.DeferredCandidates), len(adapted.NormalizedDelta.CandidateFacts))
	}
	candidate := adapted.NormalizedDelta.CandidateFacts[0]
	if candidate.SourceRelation != RelationInvokes || len(candidate.Facts) != 0 || len(candidate.Evidence) != 0 {
		t.Fatalf("unmapped candidate crossed authority boundary: %#v", candidate)
	}
}
func TestDeferredCandidateTamperFailsReconcileWithoutWrite(t *testing.T) {
	observed := adaptGeneratedBillingSource(t, generatedBillingSource(t),
		ambiguousGeneratedBillingRegistry(t), generatedBillingPolicy(t))
	if len(observed.DeferredCandidates) != 1 {
		t.Fatalf("deferred candidates = %d, want one", len(observed.DeferredCandidates))
	}
	before := irSnapshot(observed.IR)
	observed.DeferredCandidates[0].Reason = "tampered candidate reason"

	reconcile := ReconcileSemantic(observed, observed.IR, observed.SourceDigest, observed.PolicyDigest,
		observed.ToolchainDigest, observed.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("tampered deferred candidate reconcile = %#v, want invalid no-write", reconcile)
	}
	if got := irSnapshot(observed.IR); got != before {
		t.Fatalf("reconcile mutated IR: before=%q after=%q", before, got)
	}
}
