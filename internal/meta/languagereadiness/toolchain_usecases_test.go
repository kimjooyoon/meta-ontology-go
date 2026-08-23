package languagereadiness

import "testing"

func TestToolchainUseCasesRequireDynamicReceipt(t *testing.T) {
	obligation := Obligation{ID: "TOOLCHAIN-EXECUTABLE-USE-CASES",
		Area: areaToolchain, ProofChoice: "COHERENCE", ConceptID: toolchainUseCasesConcept}
	concept := conceptEvidence{ID: toolchainUseCasesConcept, Stage: "OPERATING",
		CodeBindings: []string{"internal/meta/languagereadiness/toolchainusecases"},
		MetricBindings: []string{"gooo.metric.toolchain.executable-use-cases-readiness-bps.v1"},
		UseCases: []useCaseEvidence{{ID: "canonical-and-negative-replay",
			Trigger: "artifact", ExpectedOutcome: "3_OF_3"}}}
	without := evaluateObligation(obligation, []conceptEvidence{concept}, evidenceDigests{})
	if without.Status != "NOT_SATISFIED" || without.Reason != "EXECUTABLE_USE_CASE_RECEIPT_REQUIRED" {
		t.Fatalf("without evidence = %#v", without)
	}
	with := evaluateObligation(obligation, []conceptEvidence{concept},
		evidenceDigests{useCases: "sha256:receipt"})
	if with.Status != "SATISFIED" || with.Reason != "CONCEPT_CONFORMANCE_EXPLICIT" {
		t.Fatalf("with evidence = %#v", with)
	}
}
