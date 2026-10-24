package actionability

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metabinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestDeclaredOperationWithoutConsumerRemainsActionabilityDebt(t *testing.T) {
	metrics := metricsDocument{CommitSHA: strings.Repeat("a", 40), Repository: "example/controlled-fixture",
		Meta: sourcepolicy.Report{Indicators: []sourcepolicy.Indicator{
			{Operation: sourcepolicy.OperationSplitGooo, Blocking: true},
			{Operation: sourcepolicy.OperationSplitGo, Blocking: true},
		}}}
	bindings := metabinding.Report{Witnesses: []metabinding.Witness{
		{Operation: "split-gooo-sections", Bound: true, ProofChoice: "foundation", Registry: "generation"},
		{Operation: "split-go-declarations", Bound: true, ProofChoice: "foundation", Registry: "generation"},
	}}
	report, err := Build(filepath.Join("..", "..", ".."), input{metrics: metrics, binding: bindings})
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "IMPROVE" || report.SelectedOperation != "split-gooo-sections" ||
		report.Summary.RequiredOperations != 2 || report.Summary.ExecutableOperations != 1 ||
		report.Summary.UnactionableIndicators != 1 || report.Summary.OperationCoverageBasisPoints != 5000 {
		t.Fatalf("missing consumer disrupted independent actionability: %+v", report)
	}
	for _, witness := range report.Operations {
		if witness.Operation == "split-gooo-sections" {
			if !witness.MetaBound || witness.Executable || witness.Status != "EXECUTOR_MISSING" {
				t.Fatalf("declaration became an execution claim: %+v", witness)
			}
			return
		}
	}
	t.Fatal("declared operation disappeared from denominator")
}
