package generation

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestGeneratedOwnershipApplicabilityKeepsItsOriginalMetricShape(t *testing.T) {
	report, err := sourcepolicy.Evaluate(sourcepolicy.Default(), []sourcepolicy.Observation{{Subject: "artifact.go", Dimension: sourcepolicy.DimensionGoFileLines, Value: 100, GeneratorOwned: true}})
	if err != nil || len(report.Indicators) != 1 {
		t.Fatal(report, err)
	}
	indicator := report.Indicators[0]
	if !validIndicatorApplicability(indicator) || indicator.Value != 100 || indicator.IsDriverCandidate() {
		t.Fatal(indicator)
	}
	for _, change := range []func(*sourcepolicy.Indicator){
		func(i *sourcepolicy.Indicator) {
			i.ApplicabilityReason = sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable
		},
		func(i *sourcepolicy.Indicator) { i.Operation = sourcepolicy.OperationObserve },
		func(i *sourcepolicy.Indicator) { i.Subject = "." },
		func(i *sourcepolicy.Indicator) { i.Blocking = true },
	} {
		candidate := indicator
		change(&candidate)
		if validIndicatorApplicability(candidate) {
			t.Fatal("changed metric identity was accepted", candidate)
		}
	}
}
