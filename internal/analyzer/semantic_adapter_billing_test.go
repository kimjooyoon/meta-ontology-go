package analyzer

import (
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestSemanticAdapterBillingFixtureUsesOnlyRegisteredMappings(t *testing.T) {
	analysis := analyzeBilling(t)
	policy := billingPolicy(t, RelationUses)
	adapted := adaptBilling(t, analysis, policy)

	if got := len(adapted.IR.Graph.DeterministicFacts()); got != 2 {
		t.Fatalf("deterministic facts = %d, want 2", got)
	}
	if got := len(adapted.IR.Graph.Candidates()); got != 0 {
		t.Fatalf("candidates in semantic graph = %d, want 0", got)
	}
	if len(adapted.DeferredFacts) != 1 || adapted.DeferredFacts[0].Relation != RelationGenerates {
		t.Fatalf("unmapped generates fact was not deferred: %#v", adapted.DeferredFacts)
	}
	if len(adapted.DeferredCandidates) != 1 || adapted.DeferredCandidates[0].Relation != RelationInvokes {
		t.Fatalf("unmapped invokes candidate was not preserved: %#v", adapted.DeferredCandidates)
	}
	for _, fact := range adapted.IR.Graph.DeterministicFacts() {
		if fact.Predicate != semantic.Used {
			t.Errorf("unregistered PROV predicate emitted: %s", fact.Predicate)
		}
	}
	if got := len(adapted.IR.Evidence()); got != 2 {
		t.Fatalf("evidence = %d, want 2", got)
	}
	if err := adapted.IR.Validate(); err != nil {
		t.Fatalf("adapted billing IR invalid: %v", err)
	}
}

func TestSemanticAdapterBillingFixtureStableIDsSpansAndEvidence(t *testing.T) {
	analysis := analyzeBilling(t)
	policy := billingPolicy(t, RelationUses, RelationGenerates)
	first := adaptBilling(t, analysis, policy)
	second := adaptBilling(t, analysis, policy)

	if first.IR.StableHash() != second.IR.StableHash() {
		t.Fatal("semantic hash changed across repeated adaptation")
	}
	if first.IR.EvidenceHash() != second.IR.EvidenceHash() || first.IR.ProvenanceHash() != second.IR.ProvenanceHash() {
		t.Fatal("evidence or provenance hash changed across repeated adaptation")
	}
	if !reflect.DeepEqual(first.IR.Graph.Nodes(), second.IR.Graph.Nodes()) {
		t.Fatal("registered semantic nodes changed across repeated adaptation")
	}
	for _, fact := range first.IR.Graph.DeterministicFacts() {
		if fact.Span.File == "" || fact.Span.End.Offset <= fact.Span.Start.Offset {
			t.Fatalf("fact lost source span: %#v", fact)
		}
	}
}

func analyzeBilling(t *testing.T) Result {
	t.Helper()
	source, err := os.ReadFile("testdata/registered.go")
	if err != nil {
		t.Fatalf("read billing fixture: %v", err)
	}
	registry := NewRegistry()
	registry.MustRegister(Registration{
		Ref:  SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"},
		Kind: KindActivity, Identity: NewIdentity("fraud", "fraud://activity/check"),
	})
	registry.MustRegister(Registration{
		Ref:  SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"},
		Kind: KindActivity, Identity: NewIdentity("security", "security://activity/check"),
	})
	result, err := AnalyzeSource("testdata/registered.go", source, registry)
	if err != nil {
		t.Fatalf("analyze billing fixture: %v", err)
	}
	return result
}

func billingPolicy(t *testing.T, relations ...Relation) MappingPolicy {
	t.Helper()
	policy, err := NewMappingPolicy(CurrentSemanticAdapterPolicy)
	if err != nil {
		t.Fatal(err)
	}
	for _, relation := range relations {
		mapping := RelationMapping{
			Source: relation, Predicate: semantic.Used,
			SourceSubjectKind: semantic.Activity, SourceObjectKind: semantic.Entity,
		}
		if relation == RelationGenerates {
			mapping.Predicate = semantic.WasGeneratedBy
			mapping.Reverse = true
		}
		if err := policy.Register(mapping); err != nil {
			t.Fatalf("register %s mapping: %v", relation, err)
		}
	}
	return policy
}

func adaptBilling(t *testing.T, analysis Result, policy MappingPolicy) SemanticAdapterResult {
	t.Helper()
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	adapted, err := AdaptSemantic(SemanticAdapterInput{
		Base: base, Analysis: analysis, Policy: policy,
		Producer:     semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence,
		SourceDigest: semantic.StableHash([]byte("billing-fixture/v1")),
	})
	if err != nil {
		t.Fatalf("adapt billing fixture: %v", err)
	}
	return adapted
}
