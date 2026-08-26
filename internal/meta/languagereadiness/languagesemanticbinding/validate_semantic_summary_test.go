package languagesemanticbinding

import "testing"

func TestVersionedDenominatorsMatchActivatedCorpus(t *testing.T) {
	if syntaxCaseDenominator != 26 || syntaxValidSourceDenominator != 23 || syntaxInvalidCaseDenominator != 3 || syntaxGoooLineDenominator != 356 {
		t.Fatal("syntax denominator must match the activated 26-case, 23-source, 356-line corpus")
	}
	if semanticCaseDenominator != 28 || semanticSourceDenominator != 23 {
		t.Fatal("semantic denominator must match the activated 28-case, 23-source corpus")
	}
}

func exactSemanticSummary() semanticSummary {
	return semanticSummary{
		Satisfied:                  semanticCaseDenominator,
		Total:                      semanticCaseDenominator,
		Executed:                   semanticCaseDenominator,
		ReadinessBPS:               10000,
		SourceModels:               semanticSourceDenominator,
		NormalizedIRs:              semanticSourceDenominator,
		SemanticReplays:            semanticSourceDenominator,
		ProvenanceReplays:          semanticSourceDenominator,
		EvidenceReplays:            semanticSourceDenominator,
		PresentationLaws:           1,
		CandidateAuthorityLaws:     1,
		DeterministicAuthorityLaws: 1,
		UpstreamRejections:         2,
	}
}

func TestValidateSemanticSummaryAcceptsVersionedDenominator(t *testing.T) {
	if err := validateSemanticSummary(exactSemanticSummary()); err != nil {
		t.Fatalf("versioned semantic denominator must remain consumable: %v", err)
	}
}

func TestValidateSemanticSummaryRejectsStaleDenominator(t *testing.T) {
	summary := exactSemanticSummary()
	summary.Satisfied, summary.Total, summary.Executed = semanticCaseDenominator-1, semanticCaseDenominator-1, semanticCaseDenominator-1
	if err := validateSemanticSummary(summary); err == nil {
		t.Fatal("stale semantic denominator must fail closed")
	}
}

func TestValidateSemanticEvidenceRejectsStaleCaseDenominator(t *testing.T) {
	value := semanticArtifact{Cases: make([]semanticCase, semanticCaseDenominator-1)}
	if err := validateSemanticEvidence(value, nil); err == nil {
		t.Fatal("stale semantic evidence denominator must fail closed")
	}
}
