package pressureindependence

// EvaluateBaseline is a deliberately separate typed-config/full-suite
// comparison. It observes every declared pressure and never reuses the
// pressure-independence evaluator's selection helpers.
func EvaluateBaseline(input Input) BaselineResult {
	result := BaselineResult{FullSuite: true, LocalizedIDs: []string{}}
	result.WorkUnits = baselineWork(input)
	result.CostReceipt = baselineReceipt(input, result.WorkUnits)
	if baselineMissing(input) {
		return baselineUnknown(result, input, ReasonRequiredInputMissing)
	}
	if baselineStale(input) {
		return baselineUnknown(result, input, ReasonStaleDigest)
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		if prior, exists := records[record.PressureID]; exists {
			if prior == record {
				return baselineFail(result, input, ReasonDuplicatePressureID)
			}
			return baselineFail(result, input, ReasonConflictingGroupBinding)
		}
		records[record.PressureID] = record
	}
	groups := make(map[string]struct{})
	applicabilityRule := ""
	for _, id := range input.RequiredPressureIDs {
		record, exists := records[id]
		if !exists {
			return baselineFail(result, input, ReasonCatalogMismatch)
		}
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return baselineUnknown(result, input, ReasonApplicabilityUnproven)
		}
		if applicabilityRule == "" {
			applicabilityRule = record.ApplicabilityRuleID
		} else if applicabilityRule != record.ApplicabilityRuleID {
			return baselineUnknown(result, input, ReasonInputAmbiguous)
		}
		groups[record.IndependenceGroupID] = struct{}{}
	}
	if uint64(len(groups)) < input.MinimumIndependent ||
		uint64(len(groups)) < baselineK(input.RequestedK) {
		return baselineUnknown(result, input, ReasonIndependentGroupShortfall)
	}
	result.Decision, result.Reason = DecisionPass, ReasonNone
	return result
}
func Compare(input Input) Comparison {
	oracle := Evaluate(input)
	baseline := EvaluateBaseline(input)
	localized := oracle.UnknownIDs
	if oracle.Decision == DecisionPass {
		localized = nil
	}
	comparison := Comparison{
		Oracle: oracle, Baseline: baseline, ResearchWorkUnits: oracle.CostReceipt.WorkUnits,
		BaselineWorkUnits: baseline.WorkUnits,
		OutcomeMatch:      oracle.Decision == baseline.Decision,
		ReasonMatch:       oracle.Reason == baseline.Reason,
		LocalizationMatch: equalStrings(localized, baseline.LocalizedIDs),
	}
	comparison.ResearchBudgetOK = withinResearchBudget(comparison.ResearchWorkUnits, comparison.BaselineWorkUnits)
	if comparison.OutcomeMatch && comparison.ReasonMatch && comparison.LocalizationMatch && comparison.ResearchBudgetOK {
		comparison.Finding = NoUniqueBenefit
	} else {
		comparison.Finding = UniqueBenefitNotEstablished
	}
	return comparison
}
