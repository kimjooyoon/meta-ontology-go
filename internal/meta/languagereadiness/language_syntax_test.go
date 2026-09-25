package languagereadiness

import "testing"

func TestLanguageSyntaxRequiresDynamicReceipt(t *testing.T) {
	obligation := Obligation{ID: "LANGUAGE-SYNTAX-ROUNDTRIP",
		Area: areaLanguage, ProofChoice: "FOUNDATION", ConceptID: languageSyntaxConcept}
	concept := conceptEvidence{ID: languageSyntaxConcept, Stage: "OPERATING",
		CodeBindings:   []string{"internal/meta/languagereadiness/languagesyntax"},
		MetricBindings: []string{"gooo.metric.language.syntax-roundtrip-readiness-bps.v1"},
		UseCases: []useCaseEvidence{{ID: "complete-syntax-corpus-replay",
			Trigger: "15 cases", ExpectedOutcome: "15_OF_15"}}}
	without := evaluateObligation(obligation, []conceptEvidence{concept}, evidenceDigests{})
	if without.Status != "NOT_SATISFIED" || without.Reason != "LANGUAGE_SYNTAX_ROUNDTRIP_RECEIPT_REQUIRED" {
		t.Fatalf("without evidence = %#v", without)
	}
	with := evaluateObligation(obligation, []conceptEvidence{concept},
		evidenceDigests{syntax: "sha256:receipt"})
	if with.Status != "SATISFIED" || with.Reason != "CONCEPT_CONFORMANCE_EXPLICIT" {
		t.Fatalf("with evidence = %#v", with)
	}
}
