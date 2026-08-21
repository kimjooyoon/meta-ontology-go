package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"os"
	"testing"
)

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
