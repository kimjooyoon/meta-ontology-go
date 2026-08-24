package languagesemanticbinding

import "testing"

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
	if err := validateSemanticEvidence(value, nil); err == nil || err.Error() != "semantic case denominator mismatch" {
		t.Fatalf("stale semantic evidence denominator must fail closed, got %v", err)
	}
}
