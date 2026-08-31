package operationconformance

func Evaluate(contractRaw []byte, evidence SplitGoEvidence) Report {
	contract, contractErr := EvaluateContract(contractRaw)
	evidenceDigest := digestValue(evidence)
	observations := make([]IndicatorObservation, 0, len(fixedIndicators))
	summary := Summary{Total: len(fixedIndicators), IndependentImplementationCount: 1}
	for _, definition := range fixedIndicators {
		decision := DecisionUnknown
		if contractErr == nil && validSubject(evidence) {
			decision = observeIndicator(definition.ID, evidence)
		}
		observations = append(observations, observation(definition, decision, evidenceDigest))
		switch decision {
		case DecisionPass:
			summary.PassCount++
		case DecisionFail:
			summary.FailCount++
		case DecisionUnknown:
			summary.UnknownCount++
		}
	}
	summary.EvaluatedIndicatorCount = summary.Total - summary.UnknownCount
	summary.RuntimeObservedIndicatorCount = summary.EvaluatedIndicatorCount
	decision, reason, resolution := DecisionBlock, "SPLIT_GO_EVIDENCE_UNKNOWN", "LOWER_RESOLUTION"
	if contractErr != nil {
		reason = "SPLIT_GO_CONTRACT_UNKNOWN"
	} else if summary.UnknownCount == 0 && summary.FailCount != 0 {
		reason, resolution = "SPLIT_GO_INDICATOR_FAILED", "EXACT"
	} else if summary.PassCount == summary.Total {
		decision, reason, resolution = DecisionPass, "SPLIT_GO_CONFORMANT", "EXACT"
	}
	report := Report{Schema: ReportSchema, ContractID: ContractID, OperationID: OperationID,
		Decision: decision, Reason: reason, Resolution: resolution,
		AssuranceGrade: "E1_SEPARATE_JUDGE_SAME_REPOSITORY", MetaOperation: "evaluate-split-go-behavior",
		Contract: contract, Evidence: evidence, EvidenceDigest: evidenceDigest,
		Summary: summary, Indicators: observations, Counterexamples: counterexamples(observations), RepositoryWrites: 0,
		RepositoryMutationAuthorized: false}
	report.Proofs = buildProofs(report)
	return seal(report)
}

func counterexamples(observations []IndicatorObservation) []IndicatorCounterexample {
	result := make([]IndicatorCounterexample, 0)
	for _, observation := range observations {
		if observation.Decision == DecisionPass {
			continue
		}
		result = append(result, IndicatorCounterexample{IndicatorID: observation.ID,
			RuleID: observation.RuleID, Observed: observation.Value, Expected: observation.Target,
			Decision: observation.Decision, EvidenceDigest: observation.ObservationDigest})
	}
	return result
}

func validSubject(value SplitGoEvidence) bool {
	return value.OperationID == OperationID && validSHA(value.ExpectedHeadSHA)
}
