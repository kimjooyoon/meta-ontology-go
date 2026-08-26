package verify

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func directoryIndicatorViolation(indicator sourcepolicy.Indicator) (Violation, bool) {
	switch indicator.MetricID {
	case sourcepolicy.DimensionDirectEntries:
		return Violation{
			Path: indicator.Subject, Rule: "directory direct entries",
			Actual: indicator.Value, Limit: indicator.Limit,
			Detail: "too many direct children",
		}, true
	case sourcepolicy.DimensionDirectoryKinds:
		return Violation{
			Path: indicator.Subject, Rule: "directory mixed entries",
			Actual: indicator.Value, Limit: indicator.Limit,
			Detail: "must contain either files or folders, not both",
		}, true
	default:
		return Violation{}, false
	}
}
