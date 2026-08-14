package analyzer

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestSemanticAdapterRejectsUnknownCandidateRelationWithoutMutation(t *testing.T) {
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
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
				Relation: Relation("free-form"), Options: []Identity{
					NewIdentity("billing", "billing://entity/order"),
				}, Span: testSpan(), Reason: "unknown relation",
			}},
		},
	}
	_, err := AdaptSemantic(SemanticAdapterInput{
		Base: base, Analysis: analysis, Policy: billingPolicy(t, RelationUses),
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		SourceDigest: semantic.StableHash([]byte("unknown-candidate-relation")),
	})
	assertAdapterCode(t, err, AdapterUnknownRelation)
	assertSnapshot(t, base, before)
}

func TestSemanticAdapterRejectsUnknownCandidateEndpointWithoutMutation(t *testing.T) {
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	analysis := Result{
		Registrations: billingRegistrations(),
		Delta: SemanticDelta{Candidates: []Candidate{
			candidateWithOption("billing://entity/order"),
			candidateWithOption("billing://entity/missing"),
		}},
	}
	_, err := AdaptSemantic(SemanticAdapterInput{
		Base: base, Analysis: analysis, Policy: billingPolicy(t, RelationUses),
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		SourceDigest: semantic.StableHash([]byte("unknown-candidate-endpoint")),
	})
	assertAdapterCode(t, err, AdapterUnknownEndpoint)
	assertSnapshot(t, base, before)
}

func candidateWithOption(object string) Candidate {
	return Candidate{
		Subject: NewIdentity("billing", "billing://activity/pay-order"), Relation: RelationUses,
		Options: []Identity{NewIdentity("billing", object)}, Span: testSpan(), Reason: "ambiguous endpoint",
	}
}
