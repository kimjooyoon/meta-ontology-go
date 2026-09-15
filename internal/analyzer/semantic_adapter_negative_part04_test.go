package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestSemanticAdapterShadowedEvidencePreservesDistinctSpansAndOrder(t *testing.T) {
	base := shadowedCandidateBase(t)
	policy := billingPolicy(t, RelationUses)
	candidates := []Candidate{
		shadowedCandidate(20, 32),
		shadowedCandidate(40, 52),
	}
	first := adaptShadowedCandidates(t, base, policy, candidates)
	second := adaptShadowedCandidates(t, base, policy, []Candidate{candidates[1], candidates[0]})
	if first.ShadowedCandidateEvidenceHash() != second.ShadowedCandidateEvidenceHash() {
		t.Fatal("shadowed evidence hash changed with candidate input order")
	}
	if got := len(first.ShadowedCandidateEvidence); got != 2 {
		t.Fatalf("shadowed evidence = %d, want 2", got)
	}
	if first.ShadowedCandidateEvidence[0].ID == first.ShadowedCandidateEvidence[1].ID {
		t.Fatal("distinct source spans reused one shadowed evidence ID")
	}
	if first.ShadowedCandidateEvidence[0].Span.Start.Offset == first.ShadowedCandidateEvidence[1].Span.Start.Offset {
		t.Fatal("distinct shadowed evidence spans collapsed")
	}
	if got := len(first.IR.Evidence()); got != 0 {
		t.Fatalf("shadowed observations entered authoritative evidence: %d", got)
	}
}
func shadowedCandidateBase(t *testing.T) semantic.IR {
	t.Helper()
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	activity, err := semantic.NewActivity(semantic.MustIdentity("billing://activity/pay-order"), semantic.Namespace("billing"), "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	entity, err := semantic.NewEntity(semantic.MustIdentity("billing://entity/order"), semantic.Namespace("billing"), "Order")
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range []semantic.Node{activity, entity} {
		if err := base.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := base.AddFact(semantic.NewUsedFact(activity.ID, entity.ID)); err != nil {
		t.Fatal(err)
	}
	return base
}
func shadowedCandidate(start, end int) Candidate {
	return Candidate{
		Subject: NewIdentity("billing", "billing://activity/pay-order"), Relation: RelationUses,
		Options: []Identity{NewIdentity("billing", "billing://entity/order")},
		Span: Span{Filename: "billing.go", Start: Position{Offset: start, Line: 2, Column: 1},
			End: Position{Offset: end, Line: 2, Column: end - start + 1}},
		Reason: "ambiguous implementation reference",
	}
}
func adaptShadowedCandidates(t *testing.T, base semantic.IR, policy MappingPolicy, candidates []Candidate) SemanticAdapterResult {
	t.Helper()
	adapted, err := AdaptSemantic(SemanticAdapterInput{
		Base: base, Analysis: Result{Delta: SemanticDelta{Candidates: candidates}}, Policy: policy,
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		SourceDigest: semantic.StableHash([]byte("shadowed-candidate")),
	})
	if err != nil {
		t.Fatal(err)
	}
	return adapted
}
