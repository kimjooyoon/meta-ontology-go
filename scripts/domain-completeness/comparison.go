package main

import "github.com/kimjooyoon/meta-ontology-go/internal/domaincompleteness"

func validComparisonCounts(dimension Dimension) bool {
	return dimension.Denominator > 0 && dimension.Numerator >= 0 && dimension.Numerator <= dimension.Denominator &&
		dimension.UnknownUnits >= 0 && dimension.RefutedUnits >= 0
}

func compareDimension(current, prior Dimension) DimensionDelta {
	result := DimensionDelta{ID: current.ID, Status: "UNKNOWN_UNRESOLVED_EVIDENCE",
		Denominator: current.Denominator, BaselineStatus: prior.Status, CurrentStatus: current.Status}
	if domaincompleteness.CanCompareDomainCompletenessAxes(
		int64(prior.UnknownUnits), int64(current.UnknownUnits), int64(prior.RefutedUnits), int64(current.RefutedUnits),
		prior.Status, current.Status,
	) {
		result.Status = "COMPARABLE"
		delta := current.Numerator - prior.Numerator
		result.NumeratorDelta = &delta
	}
	return result
}
