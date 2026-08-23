package toolchainformatfix

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func TestUnknownRegistryLowersFormatFixResolution(t *testing.T) {
	executor := &fakeExecutor{}
	report := Evaluate(Input{ExpectedHeadSHA: testHead,
		ConceptArtifact: languageconcept.BuildArtifact(os.DirFS("../../../..")),
		RegistryRaw: []byte("{\"schema\":\"unknown\"}"), Executor: executor})
	if err := Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionClosed || report.Resolution != ResolutionLower ||
		report.ReasonCode != "FORMAT_FIX_EVIDENCE_UNKNOWN" ||
		report.Summary.Unresolved != FixedTotal || executor.calls != 0 {
		t.Fatalf("report = %#v calls=%d", report, executor.calls)
	}
}

func TestUnknownTopDecisionIsRejected(t *testing.T) {
	report := Evaluate(Input{ExpectedHeadSHA: testHead,
		ConceptArtifact: languageconcept.BuildArtifact(os.DirFS("../../../..")),
		RegistryRaw: registryFixture(t), Executor: &fakeExecutor{}})
	report.Decision = "FIXED_POINT"
	report = seal(report)
	if Validate(report, testHead) == nil {
		t.Fatal("unknown top decision accepted")
	}
}
