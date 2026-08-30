package entityfields

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/entityfieldsv1"
)

func TestEntityFieldsReportClosesExactTwelveCells(t *testing.T) {
	observation, err := entityfieldsv1.Observe("fixture.gooo", entityfieldsv1.CanonicalSource)
	if err != nil {
		t.Fatal(err)
	}
	report := Evaluate(observation)
	if err := Verify(observation, report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.CellsTotal != 12 || report.Summary.ClosedCells != 12 || report.ActivityCount != 12 || report.BindingCount != 12 {
		t.Fatalf("report = %#v", report)
	}
}

func TestEntityFieldsCounterexamplesAreTypedAndNonPartial(t *testing.T) {
	for _, counterexample := range FixedCounterexamples() {
		if counterexample.Decision == DecisionRefuted && counterexample.Resolution != ResolutionExact {
			t.Fatalf("counterexample = %#v", counterexample)
		}
		if counterexample.Decision == DecisionRefuted && counterexample.PartialOutput {
			t.Fatalf("refuted case permits partial output: %#v", counterexample)
		}
	}
	missing := FixedCounterexamples()[5]
	if missing.Decision != "UNKNOWN" || missing.Unknown == nil || missing.Unknown.UnknownClass != "DIRECT_MISSING" || len(missing.Unknown.BlockedBy) != 0 {
		t.Fatalf("unknown shape = %#v", missing)
	}
	if !strings.Contains(missing.Unknown.NextOperation, "RESTORE") {
		t.Fatalf("unknown next operation = %#v", missing.Unknown)
	}
}
