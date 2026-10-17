package toolchainformatfix

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func TestEvaluateProvesExecutablePlansAndFixedPoints(t *testing.T) {
	executor := &fakeExecutor{}
	report := Evaluate(Input{ExpectedHeadSHA: testHead,
		ConceptArtifact: languageconcept.BuildArtifact(os.DirFS("../../../..")),
		RegistryRaw:     registryFixture(t), Executor: executor})
	if err := Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact ||
		report.Summary.Satisfied != FixedTotal || report.Summary.Invocations != ExpectedRuns ||
		report.Summary.StructuredPlans != ExpectedPlans ||
		report.Summary.InMemoryApplications != 1 || report.Summary.FixedPoints != 2 ||
		report.RepositoryWrites != 0 {
		t.Fatalf("report = %#v", report)
	}
	for _, result := range report.Cases {
		if result.First.PeakRSSKiB <= 0 || result.Replay.PeakRSSKiB <= 0 ||
			result.First.PeakRSSKiB == result.Replay.PeakRSSKiB || !result.ReplayMatched {
			t.Fatal("runner observations were lost or changed format/fix replay equality", result)
		}
	}
}
