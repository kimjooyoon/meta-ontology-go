package languagereadiness

import "testing"

func TestLanguagePackageRuntimeRequiresDynamicReceipt(t *testing.T) {
	obligation := Obligation{ID: "LANGUAGE-PACKAGE-RUNTIME",
		Area: areaLanguage, ProofChoice: "FOUNDATION", ConceptID: packageRuntimeConcept}
	concept := conceptEvidence{ID: packageRuntimeConcept, Stage: "OPERATING",
		CodeBindings: []string{"internal/meta/languagereadiness/languagepackageruntime"},
		MetricBindings: []string{"gooo.metric.language.package-runtime-readiness-bps.v1"},
		UseCases: []useCaseEvidence{{ID: "fixed-package-runtime-corpus",
			Trigger: "18 cases", ExpectedOutcome: "18_OF_18"}}}
	without := evaluateObligation(obligation, []conceptEvidence{concept}, evidenceDigests{})
	if without.Status != "NOT_SATISFIED" || without.Reason != "LANGUAGE_PACKAGE_RUNTIME_RECEIPT_REQUIRED" {
		t.Fatalf("without evidence = %#v", without)
	}
	with := evaluateObligation(obligation, []conceptEvidence{concept},
		evidenceDigests{packageRuntime: "sha256:receipt"})
	if with.Status != "SATISFIED" || with.Reason != "CONCEPT_CONFORMANCE_EXPLICIT" {
		t.Fatalf("with evidence = %#v", with)
	}
}
