package languagereadiness

import "testing"

func TestToolchainConformanceRequiresExactDynamicReceipt(t *testing.T) {
	obligation := Obligation{ID: "TOOLCHAIN-CONFORMANCE", Area: areaToolchain,
		ProofChoice: "REGRESSION", ConceptID: toolchainConformanceConcept}
	concept := conceptEvidence{ID: toolchainConformanceConcept, Stage: "OPERATING",
		CodeBindings: []string{
			"internal/meta/languagereadiness/toolchainconformance",
		},
		MetricBindings: []string{"gooo.metric.toolchain.conformance-readiness-bps.v1"},
		UseCases: []useCaseEvidence{{ID: "same-head-surface-closure",
			Trigger: "9 surfaces", ExpectedOutcome: "9_OF_9"}}}
	without := evaluateObligation(obligation, []conceptEvidence{concept}, evidenceDigests{})
	if without.Status != "NOT_SATISFIED" ||
		without.Reason != "TOOLCHAIN_CONFORMANCE_RECEIPT_REQUIRED" {
		t.Fatalf("without evidence = %#v", without)
	}
	with := evaluateObligation(obligation, []conceptEvidence{concept},
		evidenceDigests{toolchainConformance: "sha256:receipt"})
	if with.Status != "SATISFIED" || with.Reason != "CONCEPT_CONFORMANCE_EXPLICIT" {
		t.Fatalf("with evidence = %#v", with)
	}
}
