package analyzer

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

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
