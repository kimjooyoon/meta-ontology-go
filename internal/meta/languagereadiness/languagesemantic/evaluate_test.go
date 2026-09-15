package languagesemantic

import "testing"

func TestValidateSyntaxReceiptRejectsUnknownDecision(t *testing.T) {
	head := "0123456789012345678901234567890123456789"
	receipt := syntaxReceipt{Schema: "gooo/language-syntax-roundtrip/v1", Decision: "UNKNOWN", Resolution: "EXACT", Source: syntaxSource{ExpectedHeadSHA: head, ObservationKnown: true, ConceptBound: true}}
	if err := validateSyntaxReceipt(receipt, head, "", nil); err == nil {
		t.Fatal("unknown syntax decision must lower semantic resolution")
	}
}

func TestSemanticSourceProjectionUsesSemanticRegistry(t *testing.T) {
	registry := Registry{Cases: []Definition{
		{ID: "source-a", Kind: CaseSource, Path: "a.gooo"},
		{ID: "source-b", Kind: CaseSource, Path: "b.gooo"},
		{ID: "law", Kind: CaseLaw, Law: "PRESENTATION_INVARIANCE"},
	}}
	receipt := syntaxReceipt{Source: syntaxSource{
		PackageUnits: []syntaxPackageUnit{{Members: []string{"package.gooo"}}},
		GoooFiles:    []GoooFile{{Path: "a.gooo"}, {Path: "package.gooo"}, {Path: "unknown.gooo"}},
	}}
	paths := semanticSourcePaths(registry, receipt)
	if len(paths) != 1 || paths[0] != "a.gooo" {
		t.Fatalf("semantic source paths = %#v", paths)
	}
}

func TestRegistryRejectsUnknownLaw(t *testing.T) {
	registry := Registry{Schema: RegistrySchema, Version: "test", Cases: make([]Definition, FixedTotal)}
	for index := range registry.Cases {
		registry.Cases[index] = Definition{ID: string(rune('a' + index)), Kind: CaseLaw, Law: "UNKNOWN", ProofChoice: "COHERENCE", MetaOperation: "test"}
	}
	if err := registry.Validate(); err == nil {
		t.Fatal("unknown semantic law must fail closed")
	}
}
