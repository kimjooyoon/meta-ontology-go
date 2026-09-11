package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func buildWithoutRegistrationInputs(baseSHA, headSHA string, report sourcepolicy.Report) Plan {
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
	if failures := floorFailures(floors, registry); len(failures) != 0 {
		plan.Decision, plan.Reason = DecisionRejected, ReasonFloorRegression
		plan.UnknownIndicatorIDs = failures
		return finish(plan)
	}
	actionable, unknown, refuted, counterexamples := partitionIndicatorsForRegistry(indicators, registry)
	plan.RefutedIndicatorIDs = append(plan.RefutedIndicatorIDs, refuted...)
	plan.Counterexamples = append(plan.Counterexamples, counterexamples...)
	if len(unknown) != 0 {
		plan.Decision, plan.Reason = DecisionUnknown, ReasonMissingOperation
		plan.UnknownIndicatorIDs = unknown
		return finish(plan)
	}
	return attachPlanIndicatorDecisionLedger(selectActions(plan, actionable, registry), indicators)
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

func floorFailures(floors []sourcepolicy.Indicator, registry []Binding) []string {
	index, valid := registryIndex(registry)
	if !valid {
		return []string{"registry"}
	}
	result := make([]string, 0)
	for _, indicator := range floors {
		if !indicator.Satisfied {
			if _, routable := index[indicator.Operation]; routable {
				continue
			}
			result = append(result, indicatorID(indicator))
		}
	}
	return result
}
