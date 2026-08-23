package languagereadiness

import "testing"

func TestToolchainCLIRequiresExactDynamicReceipt(t *testing.T) {
	obligation := Obligation{ID: "TOOLCHAIN-CLI", Area: areaToolchain,
		ProofChoice: "FOUNDATION", ConceptID: toolchainCLIConcept}
	concept := conceptEvidence{ID: toolchainCLIConcept, Stage: "OPERATING",
		CodeBindings: []string{"internal/meta/languagereadiness/toolchaincli"},
		MetricBindings: []string{"gooo.metric.toolchain.cli-readiness-bps.v1"},
		UseCases: []useCaseEvidence{{ID: "fixed-cli-corpus", Trigger: "12 cases", ExpectedOutcome: "12_OF_12"}}}
	without := evaluateObligation(obligation, []conceptEvidence{concept}, evidenceDigests{})
	if without.Status != "NOT_SATISFIED" || without.Reason != "TOOLCHAIN_CLI_RECEIPT_REQUIRED" {
		t.Fatalf("without evidence = %#v", without)
	}
	with := evaluateObligation(obligation, []conceptEvidence{concept},
		evidenceDigests{toolchainCLI: "sha256:receipt"})
	if with.Status != "SATISFIED" || with.Reason != "CONCEPT_CONFORMANCE_EXPLICIT" {
		t.Fatalf("with evidence = %#v", with)
	}
}
