package languagereadiness

import "testing"

func TestToolchainFormatFixRequiresExactDynamicReceipt(t *testing.T) {
	obligation := Obligation{ID: "TOOLCHAIN-FORMAT-FIX", Area: areaToolchain,
		ProofChoice: "COHERENCE", ConceptID: toolchainFormatFixConcept}
	concept := conceptEvidence{ID: toolchainFormatFixConcept, Stage: "OPERATING",
		CodeBindings: []string{"internal/meta/languagereadiness/toolchainformatfix"},
		MetricBindings: []string{"gooo.metric.toolchain.format-fix-readiness-bps.v1"},
		UseCases: []useCaseEvidence{{ID: "fixed-format-fix-corpus",
			Trigger: "12 cases", ExpectedOutcome: "12_OF_12"}}}
	without := evaluateObligation(obligation, []conceptEvidence{concept}, evidenceDigests{})
	if without.Status != "NOT_SATISFIED" ||
		without.Reason != "TOOLCHAIN_FORMAT_FIX_RECEIPT_REQUIRED" {
		t.Fatalf("without evidence = %#v", without)
	}
	with := evaluateObligation(obligation, []conceptEvidence{concept},
		evidenceDigests{toolchainFormatFix: "sha256:receipt"})
	if with.Status != "SATISFIED" || with.Reason != "CONCEPT_CONFORMANCE_EXPLICIT" {
		t.Fatalf("with evidence = %#v", with)
	}
}
