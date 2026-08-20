package sourcepolicy

import "testing"

func TestEvaluateSelectsAllTrilemmaBranchesAndOperations(t *testing.T) {
	report, err := Evaluate(Default(), []Observation{
		{Subject: "large.go", Dimension: DimensionGoFileLines, Value: 76},
		{Subject: ".", Dimension: DimensionFixDelta, Value: 0},
		{Subject: "go.mod", Dimension: DimensionToolchain, Value: 1},
		{Subject: "wrapper.go:3:wrapper", Dimension: DimensionRefactorReturn, Value: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Indicators) != 4 || len(report.Actionable()) != 2 || len(report.Failed()) != 1 {
		t.Fatalf("unexpected indicator report: %#v", report)
	}
	proofs := map[ProofChoice]bool{}
	for _, indicator := range report.Indicators {
		proofs[indicator.Proof] = true
		if indicator.Consumer == "" || indicator.Operation == "" {
			t.Fatalf("metric is not connected to meta code: %#v", indicator)
		}
		if indicator.MetricID == DimensionRefactorReturn && (indicator.Blocking || indicator.Satisfied || indicator.Operation != OperationInspectWrapper) {
			t.Fatalf("refactor candidate lost its non-blocking operation: %#v", indicator)
		}
	}
	for _, proof := range []ProofChoice{ProofFoundation, ProofCoherence, ProofRegression} {
		if !proofs[proof] {
			t.Fatalf("missing trilemma proof choice %q", proof)
		}
	}
}
