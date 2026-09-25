package analyzer

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestShadowedCandidateEvidenceMirrorsNormalizedCandidate(t *testing.T) {
	base := shadowedCandidateBase(t)
	policy := billingPolicy(t, RelationUses)
	result, err := AdaptSemantic(SemanticAdapterInput{
		Base: base,
		Analysis: Result{Delta: SemanticDelta{Candidates: []Candidate{
			shadowedCandidate(20, 32),
		}}},
		Policy:          policy,
		Producer:        semantic.GoHostedCompilerID,
		EvidenceKind:    semantic.CompilerRunEvidence,
		SourceDigest:    semantic.StableHash([]byte("shadowed-candidate-binding")),
		ToolchainDigest: ToolchainDigest("shadowed-candidate-binding-toolchain"),
	})
	if err != nil {
		t.Fatal(err)
	}
	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if !reconcile.Accepted {
		t.Fatalf("untampered shadowed evidence reconcile = %#v, want accepted", reconcile)
	}
}

func TestShadowedCandidateEvidenceTamperFailsReconcileWithoutWrite(t *testing.T) {
	base := shadowedCandidateBase(t)
	policy := billingPolicy(t, RelationUses)
	result, err := AdaptSemantic(SemanticAdapterInput{
		Base: base,
		Analysis: Result{Delta: SemanticDelta{Candidates: []Candidate{
			shadowedCandidate(20, 32),
		}}},
		Policy:          policy,
		Producer:        semantic.GoHostedCompilerID,
		EvidenceKind:    semantic.CompilerRunEvidence,
		SourceDigest:    semantic.StableHash([]byte("shadowed-candidate-tamper")),
		ToolchainDigest: ToolchainDigest("shadowed-candidate-tamper-toolchain"),
	})
	if err != nil {
		t.Fatal(err)
	}
	result.ShadowedCandidateEvidence[0].Span.File = "tampered.go"
	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("tampered shadowed evidence reconcile = %#v, want invalid no-write", reconcile)
	}
}
