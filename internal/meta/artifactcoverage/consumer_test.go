package artifactcoverage

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/actionability"
)

func TestArtifactCannotSupplyMissingExecutionConsumer(t *testing.T) {
	action, observations := coverageFixture()
	const operation = "split-gooo-sections"
	action.Operations = append(action.Operations, actionability.OperationWitness{
		Operation: operation, IndicatorCount: 12, Executable: true,
	})
	for index := range observations.Artifacts {
		if strings.HasPrefix(observations.Artifacts[index].Name, "self-improvement-generation-") {
			observations.Artifacts[index].EvidenceKeys = append(
				observations.Artifacts[index].EvidenceKeys, "operation."+operation)
		}
	}
	report, err := Evaluate(filepath.Join("..", "..", ".."), action, observations)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "IMPROVE" || report.SelectedOperation != operation ||
		report.Summary.RequiredOperations != 8 || report.Summary.CanonicalOperations != 7 ||
		report.Summary.UncoveredOperations != 1 || report.Summary.CanonicalCoverageBasisPoints != 8750 {
		t.Fatalf("missing consumer was hidden: %+v", report)
	}
	for _, witness := range report.Operations {
		if witness.Operation == operation {
			if witness.Canonical || witness.Status != "ARTIFACT_BINDING_MISSING" {
				t.Fatalf("artifact supplied an unimplemented consumer: %+v", witness)
			}
			return
		}
	}
	t.Fatal("unimplemented operation disappeared from observation")
}

func TestExtractionArtifactUsesItsImplementedEvaluator(t *testing.T) {
	for _, binding := range CanonicalBindings() {
		if binding.Operation == "extract-function" {
			if binding.Executor != "bootstrap/function-extractor" ||
				binding.Evaluator != ".github/workflows/repository-projection.yml" {
				t.Fatalf("extraction consumer = %+v", binding)
			}
			return
		}
	}
	t.Fatal("implemented extraction consumer is missing")
}
