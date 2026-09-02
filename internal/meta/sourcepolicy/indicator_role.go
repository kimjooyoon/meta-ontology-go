package sourcepolicy

// IndicatorRole distinguishes an observed metric driver from an enforcement
// guardrail.  Drivers can select work without asserting correctness failure.
type IndicatorRole string

const IndicatorRoleDriver IndicatorRole = "DRIVER"

// IsLineCapMetric identifies dimensions whose thresholds select refactoring
// candidates rather than directly authorizing or rejecting a repository.
func IsLineCapMetric(metric Dimension) bool {
	switch metric {
	case DimensionGoFileLines, DimensionGoooFileLines, DimensionFunctionLines:
		return true
	default:
		return false
	}
}

// IsDriverCandidate is the single candidate predicate for line-cap metrics.
func (indicator Indicator) IsDriverCandidate() bool {
	return IsLineCapMetric(indicator.MetricID) &&
		indicator.Applicability == ApplicabilityApplicable &&
		indicator.Role == IndicatorRoleDriver &&
		!indicator.Blocking && indicator.Relation == RelationLessOrEqual && indicator.Value > indicator.Limit
}
