package languagesemantic

import "testing"

func TestValidateSyntaxReceiptRejectsUnknownDecision(t *testing.T) {
	receipt := syntaxReceipt{Schema: "gooo/language-syntax-roundtrip/v1", Decision: "UNKNOWN", Resolution: "EXACT", Source: syntaxSource{ExpectedHeadSHA: "0123456789012345678901234567890123456789", ObservationKnown: true, ConceptBound: true}}
	if err := validateSyntaxReceipt(receipt, receipt.Source.ExpectedHeadSHA); err == nil {
		t.Fatal("unknown syntax decision must lower semantic resolution")
	}
}

func TestValidateSyntaxReceiptAcceptsVersionedDenominator(t *testing.T) {
	head := "0123456789012345678901234567890123456789"
	receipt := syntaxReceipt{
		Schema:     "gooo/language-syntax-roundtrip/v1",
		Decision:   "PASS",
		Resolution: "EXACT",
		Summary: SyntaxSummary{
			Satisfied:    expectedSyntaxCases,
			Total:        expectedSyntaxCases,
			ValidCases:   expectedSyntaxValid,
			InvalidCases: expectedSyntaxInvalid,
			GoooLines:    expectedSyntaxLines,
		},
		Source: syntaxSource{
			ExpectedHeadSHA:  head,
			GoooFiles:        make([]GoooFile, expectedSyntaxFiles),
			ObservationKnown: true,
			ConceptBound:     true,
			PackageUnits:     versionedSyntaxPackages(),
		},
		Cases: versionedSyntaxCases(),
	}
	if err := validateSyntaxReceipt(receipt, head); err != nil {
		t.Fatalf("versioned syntax denominator must remain consumable: %v", err)
	}
}

func TestSemanticSourceProjectionPreservesUnknowns(t *testing.T) {
	cases := versionedSyntaxCases()
	packages := versionedSyntaxPackages()
	receipt := syntaxReceipt{Cases: cases, Source: syntaxSource{PackageUnits: packages, GoooFiles: []GoooFile{{Path: cases[0].Definition.Path}, {Path: packages[0].Members[0]}, {Path: cases[19].Definition.Path}, {Path: "unknown.gooo"}}}}
	paths := semanticSourcePaths(receipt)
	if len(paths) != 2 || paths[0] != cases[0].Definition.Path || paths[1] != "unknown.gooo" {
		t.Fatalf("semantic source paths = %#v", paths)
	}
}

func versionedSyntaxPackages() []syntaxPackageUnit {
	return expectedSyntaxPackageUnits()
}

func versionedSyntaxCases() []syntaxCase {
	cases := make([]syntaxCase, expectedSyntaxCases)
	for index := range cases {
		id := string(rune('a' + index))
		cases[index].Definition.ID, cases[index].Definition.Path = id, id+".gooo"
		cases[index].Definition.Kind = "VALID"
		if index >= expectedSyntaxValid {
			cases[index].Definition.Kind = "INVALID"
		}
	}
	return cases
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
