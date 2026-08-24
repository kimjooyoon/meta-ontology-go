package languagesemantic

import "testing"

func TestValidateSyntaxReceiptRejectsUnknownDecision(t *testing.T) {
	receipt := syntaxReceipt{
		Schema:     "gooo/language-syntax-roundtrip/v1",
		Decision:   "UNKNOWN",
		Resolution: "EXACT",
		Source: syntaxSource{
			ExpectedHeadSHA:  "0123456789012345678901234567890123456789",
			ObservationKnown: true,
			ConceptBound:     true,
		},
	}
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
		},
	}
	if err := validateSyntaxReceipt(receipt, head); err != nil {
		t.Fatalf("versioned syntax denominator must remain consumable: %v", err)
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
