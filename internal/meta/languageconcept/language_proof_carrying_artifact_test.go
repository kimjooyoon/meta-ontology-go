package languageconcept

import "testing"

func TestProofCarryingArtifactConceptBindsEvidenceAndConsumer(t *testing.T) {
	item := languageProofCarryingArtifactConcept
	if item.ID != "language-proof-carrying-artifact" || item.MetaOperation != "grant-read-only-consumption" || item.Stage != "OPERATING" {
		t.Fatalf("concept = %#v", item)
	}
	if len(item.CodeBindings) != 6 || len(item.MetricBindings) != 24 || len(item.UseCases) != 2 {
		t.Fatalf("bindings = %#v", item)
	}
	if item.UseCases[0].ExpectedOutcome != "PASS_6_OF_6_WITH_READ_ONLY_CONSUMPTION_AUTHORITY" ||
		item.UseCases[1].ExpectedOutcome != "FAIL_CLOSED_ARTIFACT_BYTES_NOT_AUTHORITY" {
		t.Fatalf("use cases = %#v", item.UseCases)
	}
}
