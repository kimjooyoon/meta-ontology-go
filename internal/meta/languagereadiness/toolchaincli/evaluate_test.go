package toolchaincli

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func TestEvaluateExecutesExactCLIContract(t *testing.T) {
	repository := os.DirFS("../../../..")
	executor := &fakeExecutor{}
	report := Evaluate(Input{ExpectedHeadSHA: testHead,
		ConceptArtifact: languageconcept.BuildArtifact(repository),
		RegistryRaw:     registryFixture(t), Executor: executor})
	if err := Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact ||
		report.Summary.Satisfied != FixedTotal || report.Summary.PositivePaths != FixedPositive ||
		report.Summary.GuardrailRejections != FixedGuardrails || report.Summary.Invocations != ExpectedRuns ||
		report.Summary.DeclaredCommands != ExpectedCommands || report.Summary.StructuredOutputs != ExpectedStructured ||
		report.Summary.LanguageOperations != ExpectedLanguageOps || report.Summary.ResourceObservations != ExpectedRuns ||
		report.Summary.PeakRSSKiB <= 0 || report.ResourceObservationMode != "RUNNER_SCOPED_NONDETERMINISTIC" ||
		report.ResourceMeasurementReplayAuthority || report.PerformanceImprovement != "UNKNOWN" || report.RepositoryWrites != 0 {
		t.Fatalf("report = %#v", report)
	}
}
