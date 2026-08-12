package analyzer

import (
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestSemanticAdapterRejectsReversedTypedEndpoint(t *testing.T) {
	analysis := Result{Registrations: []Registration{
		registration(KindEntity, "billing://entity/order", "Order"),
		registration(KindActivity, "billing://activity/pay-order", "PayOrder"),
	}, Delta: SemanticDelta{Added: []Fact{{
		Subject: NewIdentity("billing", "billing://entity/order"), Relation: RelationUses,
		Object: NewIdentity("billing", "billing://activity/pay-order"), Span: testSpan(),
	}}}}
	_, err := AdaptSemantic(SemanticAdapterInput{
		Base: semantic.NewIR("billing", semantic.Namespace("billing")), Analysis: analysis,
		Policy: billingPolicy(t, RelationUses), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, SourceDigest: semantic.StableHash([]byte("reverse")),
	})
	assertAdapterCode(t, err, AdapterEndpointKind)
}

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

func emptyPolicy(t *testing.T) MappingPolicy {
	t.Helper()
	policy, err := NewMappingPolicy(CurrentSemanticAdapterPolicy)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func adaptAnalysis(t *testing.T, analysis Result, policy MappingPolicy) SemanticAdapterResult {
	t.Helper()
	adapted, err := AdaptSemantic(SemanticAdapterInput{
		Base: semantic.NewIR("billing", semantic.Namespace("billing")), Analysis: analysis, Policy: policy,
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		SourceDigest: semantic.StableHash([]byte("candidate")),
	})
	if err != nil {
		t.Fatalf("adapt analysis: %v", err)
	}
	return adapted
}

func billingRegistrations() []Registration {
	return []Registration{
		registration(KindActivity, "billing://activity/pay-order", "PayOrder"),
		registration(KindEntity, "billing://entity/order", "Order"),
	}
}

func registration(kind SymbolKind, id, name string) Registration {
	return Registration{Kind: kind, Ref: SymbolRef{Name: name}, Identity: NewIdentity("billing", id), Span: testSpan()}
}

func testSpan() Span {
	return Span{Filename: "billing.go", Start: Position{Offset: 10, Line: 2, Column: 1}, End: Position{Offset: 20, Line: 2, Column: 11}}
}

func irSnapshot(ir semantic.IR) [5]string {
	return [5]string{ir.Canonical(), ir.SemanticCanonical(), ir.ProvenanceCanonical(), ir.StableHash(), ir.EvidenceHash()}
}

func assertSnapshot(t *testing.T, ir semantic.IR, before [5]string) {
	t.Helper()
	if got := irSnapshot(ir); got != before {
		t.Fatalf("IR changed after rejected adaptation: before=%q after=%q", before, got)
	}
}

func assertAdapterCode(t *testing.T, err error, want AdapterErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected adapter error %s", want)
	}
	var adapterErr AdapterError
	if !errors.As(err, &adapterErr) || adapterErr.Code != want {
		t.Fatalf("error = %v, want adapter code %s", err, want)
	}
}
