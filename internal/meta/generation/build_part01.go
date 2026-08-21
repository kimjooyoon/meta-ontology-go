package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func Build(baseSHA, headSHA string, report sourcepolicy.Report) Plan {
	indicators := normalizeIndicators(report.Indicators)
	registry := normalizeRegistry(DefaultRegistry())
	floors := blockingIndicators(indicators)
	input := canonicalInput{
		SchemaVersion: SchemaVersion, BaseSHA: baseSHA, HeadSHA: headSHA,
		Policy: report.Policy, Indicators: indicators, Registry: registry,
		RequestedK: requestedK, MinimumIndependent: minimumIndependent,
	}
	plan := Plan{
		SchemaVersion: SchemaVersion, BaseSHA: baseSHA, HeadSHA: headSHA,
		PolicyDigest: digestJSON(report.Policy), RegistryDigest: digestJSON(registry),
		IndicatorsDigest: digestJSON(indicators), FloorDigest: digestJSON(floors),
		InputDigest: digestJSON(input), RequestedK: requestedK,
		MinimumIndependent: minimumIndependent, ReplayProof: ProofCoherence,
		Registry: registry, NotApplicableIndicatorIDs: notApplicableIndicatorIDs(indicators),
	}
	if report.Schema != sourcepolicy.IndicatorSchema ||
		!validSHA(baseSHA) || !validSHA(headSHA) || duplicateIndicators(indicators) {
		plan.Decision, plan.Reason = DecisionUnknown, ReasonInvalidInput
		return finish(plan)
	}
	if failures := applicabilityFailures(indicators); len(failures) != 0 {
		plan.Decision, plan.Reason = DecisionUnknown, ReasonApplicabilityUnproven
		plan.UnknownIndicatorIDs = failures
		return finish(plan)
	}
	if _, valid := registryIndex(registry); !valid {
		plan.Decision, plan.Reason = DecisionUnknown, ReasonInvalidInput
		return finish(plan)
	}
	if failures := floorFailures(floors); len(failures) != 0 {
		plan.Decision, plan.Reason = DecisionRejected, ReasonFloorRegression
		plan.UnknownIndicatorIDs = failures
		return finish(plan)
	}
	actionable, unknown := partitionIndicators(indicators)
	if len(unknown) != 0 {
		plan.Decision, plan.Reason = DecisionUnknown, ReasonMissingOperation
		plan.UnknownIndicatorIDs = unknown
		return finish(plan)
	}
	return selectActions(plan, actionable, registry)
}

func blockingIndicators(indicators []sourcepolicy.Indicator) []sourcepolicy.Indicator {
	result := make([]sourcepolicy.Indicator, 0)
	for _, indicator := range indicators {
		if indicator.Applicability == sourcepolicy.ApplicabilityApplicable && indicator.Blocking {
			result = append(result, indicator)
		}
	}
	return result
}

func floorFailures(floors []sourcepolicy.Indicator) []string {
	result := make([]string, 0)
	for _, indicator := range floors {
		if !indicator.Satisfied {
			result = append(result, indicatorID(indicator))
		}
	}
	return result
}
