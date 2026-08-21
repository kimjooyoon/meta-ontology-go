package metricevidence

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestGoSplitIndicatorsNeverSelectsNotApplicableEvidence(t *testing.T) {
	var report Report
	report.Meta.Indicators = []Indicator{
		{MetricID: sourcepolicy.DimensionGoFileLines, Subject: "legacy.go",
			MetaOperation: sourcepolicy.OperationSplitGo},
		{MetricID: sourcepolicy.DimensionGoFileLines, Subject: "applicable.go",
			Applicability: sourcepolicy.ApplicabilityApplicable,
			MetaOperation: sourcepolicy.OperationSplitGo},
		{MetricID: sourcepolicy.DimensionGoFileLines, Subject: "exempt.go",
			Applicability: sourcepolicy.ApplicabilityNotApplicable,
			MetaOperation: sourcepolicy.OperationSplitGo},
	}
	selected := report.GoSplitIndicators()
	if len(selected) != 2 || !Contains(selected, "legacy.go") ||
		!Contains(selected, "applicable.go") || Contains(selected, "exempt.go") {
		t.Fatalf("invalid applicability selection: %#v", selected)
	}
}
