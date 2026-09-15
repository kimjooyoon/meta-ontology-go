package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestSemanticAdapterCandidateAndDetailStayOutsideAuthoritativeFacts(t *testing.T) {
	policy := billingPolicy(t, RelationUses)
	analysis := Result{
		Registrations: billingRegistrations(),
		Delta: SemanticDelta{
			Candidates: []Candidate{{
				Subject:  NewIdentity("billing", "billing://activity/pay-order"),
				Relation: RelationUses, Reference: "billing.Order",
				Options: []Identity{NewIdentity("billing", "billing://entity/order")},
				Span:    testSpan(), Reason: "ambiguous reference",
			}},
			ImplementationDetails: []ImplementationDetail{{
				Reference: "unknown.Call", Span: testSpan(), Reason: "unregistered semantic symbol",
			}},
		},
	}
	adapted := adaptAnalysis(t, analysis, policy)
	if len(adapted.IR.Graph.DeterministicFacts()) != 0 || len(adapted.IR.Graph.Candidates()) != 1 {
		t.Fatalf("fact views = %d deterministic, %d candidates", len(adapted.IR.Graph.DeterministicFacts()), len(adapted.IR.Graph.Candidates()))
	}
	if len(adapted.DeferredCandidates) != 1 || len(adapted.ImplementationDetails) != 1 {
		t.Fatalf("local candidate/detail views were not preserved")
	}
	if adapted.IR.Evidence()[0].Status != semantic.FactCandidate {
		t.Fatalf("candidate evidence status = %s", adapted.IR.Evidence()[0].Status)
	}
}
func TestSemanticAdapterRejectsUnknownRelationWithoutMutation(t *testing.T) {
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	analysis := Result{Delta: SemanticDelta{Added: []Fact{{
		Subject:  NewIdentity("billing", "billing://activity/pay-order"),
		Relation: Relation("free-form"), Object: NewIdentity("billing", "billing://entity/order"),
		Span: testSpan(),
	}}}}
	_, err := AdaptSemantic(SemanticAdapterInput{Base: base, Analysis: analysis, Policy: emptyPolicy(t)})
	assertAdapterCode(t, err, AdapterUnknownRelation)
	assertSnapshot(t, base, before)
}
func TestSemanticAdapterRejectsUnknownEndpointAfterEarlierValidFact(t *testing.T) {
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	analysis := Result{
		Registrations: billingRegistrations(),
		Delta: SemanticDelta{Added: []Fact{
			{Subject: NewIdentity("billing", "billing://activity/pay-order"), Relation: RelationUses, Object: NewIdentity("billing", "billing://entity/order"), Span: testSpan()},
			{Subject: NewIdentity("billing", "billing://activity/pay-order"), Relation: RelationUses, Object: NewIdentity("billing", "billing://entity/missing"), Span: testSpan()},
		}},
	}
	_, err := AdaptSemantic(SemanticAdapterInput{
		Base: base, Analysis: analysis, Policy: billingPolicy(t, RelationUses),
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		SourceDigest: semantic.StableHash([]byte("no-mutation")),
	})
	assertAdapterCode(t, err, AdapterUnknownEndpoint)
	assertSnapshot(t, base, before)
}
