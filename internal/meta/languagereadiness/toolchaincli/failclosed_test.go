package toolchaincli

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func TestUnknownRegistryLowersCLIResolution(t *testing.T) {
	executor := &fakeExecutor{}
	report := Evaluate(Input{ExpectedHeadSHA: testHead,
		ConceptArtifact: languageconcept.BuildArtifact(os.DirFS("../../../..")),
		RegistryRaw: []byte("{\"schema\":\"unknown\"}"), Executor: executor})
	if err := Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionClosed || report.Resolution != ResolutionLower ||
		report.ReasonCode != "TOOLCHAIN_CLI_EVIDENCE_UNKNOWN" ||
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

func TestRepositoryWriteFailsClosedExactly(t *testing.T) {
	report := Evaluate(Input{ExpectedHeadSHA: testHead,
		ConceptArtifact: languageconcept.BuildArtifact(os.DirFS("../../../..")),
		RegistryRaw: registryFixture(t), Executor: &fakeExecutor{writeOn: 1}})
	if err := Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionClosed || report.Resolution != ResolutionExact || report.RepositoryWrites != 1 {
		t.Fatalf("report = %#v", report)
	}
}
