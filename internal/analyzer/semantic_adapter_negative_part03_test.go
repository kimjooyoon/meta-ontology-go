package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestSemanticAdapterRetainsEvidenceForCandidateShadowedByFact(t *testing.T) {
	activityID := semantic.MustIdentity("billing://activity/pay-order")
	entityID := semantic.MustIdentity("billing://entity/order")
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	activity, err := semantic.NewActivity(activityID, semantic.Namespace("billing"), "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	entity, err := semantic.NewEntity(entityID, semantic.Namespace("billing"), "Order")
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range []semantic.Node{activity, entity} {
		if err := base.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	deterministic := semantic.NewUsedFact(activityID, entityID).WithSpan(semantic.Span{
		File: "billing.gooo", Start: semantic.Position{Offset: 1}, End: semantic.Position{Offset: 8},
	})
	if err := base.AddFact(deterministic); err != nil {
		t.Fatal(err)
	}

	policy := billingPolicy(t, RelationUses)
	adapted, err := AdaptSemantic(SemanticAdapterInput{
		Base: base,
		Analysis: Result{Delta: SemanticDelta{Candidates: []Candidate{{
			Subject:  NewIdentity("billing", activityID.String()),
			Relation: RelationUses,
			Options:  []Identity{NewIdentity("billing", entityID.String())},
			Span: Span{Filename: "billing.go", Start: Position{Offset: 20, Line: 2, Column: 1},
				End: Position{Offset: 32, Line: 2, Column: 13}},
			Reason: "ambiguous implementation reference",
		}}}},
		Policy:       policy,
		Producer:     semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence,
		SourceDigest: semantic.StableHash([]byte("shadowed-candidate")),
	})
	if err != nil {
		t.Fatalf("adapt shadowed candidate: %v", err)
	}
	if got := len(adapted.IR.Graph.DeterministicFacts()); got != 1 {
		t.Fatalf("deterministic facts = %d, want 1", got)
	}
	if got := len(adapted.IR.Graph.Candidates()); got != 0 {
		t.Fatalf("shadowed candidate entered authoritative graph: %d", got)
	}
	if got := len(adapted.IR.Evidence()); got != 0 {
		t.Fatalf("shadowed candidate evidence entered authoritative IR: %d", got)
	}
	if got := len(adapted.ShadowedCandidateEvidence); got != 1 {
		t.Fatalf("shadowed candidate evidence = %d, want 1", got)
	}
	evidence := adapted.ShadowedCandidateEvidence[0]
	if evidence.Status != semantic.FactCandidate || evidence.Fact != deterministic.Key() {
		t.Fatalf("shadowed evidence = %#v, want candidate for %v", evidence, deterministic.Key())
	}
	if evidence.Span.File != "billing.go" || evidence.Span.Start.Offset != 20 || evidence.Span.End.Offset != 32 {
		t.Fatalf("shadowed evidence span = %#v", evidence.Span)
	}
	if err := adapted.IR.Validate(); err != nil {
		t.Fatalf("adapted IR invalid after retaining deferred evidence: %v", err)
	}
}
