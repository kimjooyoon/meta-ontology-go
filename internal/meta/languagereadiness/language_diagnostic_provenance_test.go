package languagereadiness

import "testing"

func TestDiagnosticProvenanceRequiresDynamicReceipt(t *testing.T) {
	obligation := Obligation{
		ID: "LANGUAGE-DIAGNOSTIC-PROVENANCE", Area: areaLanguage,
		ProofChoice: "REGRESSION", ConceptID: diagnosticConcept,
	}
	concept := conceptEvidence{
		ID: diagnosticConcept, Stage: "OPERATING",
		CodeBindings: []string{
			"internal/meta/languagereadiness/languagediagnosticprovenance",
		},
		MetricBindings: []string{
			"gooo.metric.language.diagnostic-provenance-bps.v1",
		},
		UseCases: []useCaseEvidence{{
			ID: "trace-generated-diagnostic", Trigger: "diagnostic",
			ExpectedOutcome: "TRACE",
		}},
	}
	without := evaluateObligation(
		obligation, []conceptEvidence{concept}, evidenceDigests{},
	)
	if without.Status != "NOT_SATISFIED" ||
		without.Reason != "LANGUAGE_DIAGNOSTIC_PROVENANCE_RECEIPT_REQUIRED" {
		t.Fatalf("without evidence = %#v", without)
	}
	with := evaluateObligation(
		obligation, []conceptEvidence{concept},
		evidenceDigests{diagnostic: "sha256:receipt"},
	)
	if with.Status != "SATISFIED" {
		t.Fatalf("with evidence = %#v", with)
	}
}
