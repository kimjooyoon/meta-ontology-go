package languagesemanticbinding

import "testing"

func exactSemanticSummary() semanticSummary {
	return semanticSummary{
		Satisfied:                  20,
		Total:                      20,
		Executed:                   20,
		ReadinessBPS:               10000,
		SourceModels:               15,
		NormalizedIRs:              15,
		SemanticReplays:            15,
		ProvenanceReplays:          15,
		EvidenceReplays:            15,
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
	summary.Satisfied, summary.Total, summary.Executed = 19, 19, 19
	if err := validateSemanticSummary(summary); err == nil {
		t.Fatal("stale semantic denominator must fail closed")
	}
}
