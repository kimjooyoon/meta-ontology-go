package valueexecution

func buildIndicators(checks evidence) []Indicator {
	definitions := []struct {
		id, class, operation string
		value                bool
	}{
		{"source-parsed", "FOUNDATION", "parse-gooo-value-source", checks.sourceParsed},
		{"activity-resolved", "FOUNDATION", "resolve-gooo-activity", checks.activityResolved},
		{"program-present", "FOUNDATION", "read-computes-clause", checks.programPresent},
		{"program-semantic-binding", "COHERENCE", "bind-program-to-bidir-activity", checks.semanticBound},
		{"fingerprint-sensitive", "COHERENCE", "vary-program-and-compare-semantic-fingerprint", checks.fingerprintSensitive},
		{"registry-operation-known", "COHERENCE", "resolve-operation-registry-entry", checks.registryKnown},
		{"operand-parsed", "COHERENCE", "compile-operation-operand", checks.operandParsed},
		{"signature-supported", "COHERENCE", "check-value-signature", checks.signatureSupported},
		{"value-cases-exact", "REGRESSION", "execute-fixed-value-cases", checks.valueCasesExact},
		{"outputs-observed", "REGRESSION", "record-gooo-value-outputs", checks.outputsObserved},
		{"deterministic-replay", "REGRESSION", "replay-value-cases", checks.deterministicReplay},
		{"counterexamples-exact", "REGRESSION", "execute-fixed-counterexamples", checks.counterexamplesExact},
		{"unknown-operation-fail-closed", "REGRESSION", "reject-unknown-operation", checks.unknownFailClosed},
		{"malformed-operand-fail-closed", "REGRESSION", "reject-malformed-operand", checks.operandFailClosed},
		{"overflow-fail-closed", "REGRESSION", "reject-int64-overflow", checks.overflowFailClosed},
		{"core-ir-loss-fail-closed", "FOUNDATION", "reject-unrepresentable-core-ir-program", checks.coreIRFailClosed},
	}
	indicators := make([]Indicator, 0, len(definitions))
	for _, definition := range definitions {
		indicators = append(indicators, Indicator{
			ID: definition.id, Class: definition.class, ProofChoice: definition.class,
			MetaOperation: definition.operation, Value: boolInt(definition.value), Target: 1, Satisfied: definition.value,
		})
	}
	return indicators
}

func buildViews(indicators []Indicator) []View {
	return []View{
		buildView("USER", "VALUE_OUTPUTS", indicators, []int{2, 4, 8, 9, 10}),
		buildView("TOOL_AUTHOR", "PROGRAM_CONTRACT", indicators, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}),
		buildView("GOVERNOR", "FULL_RECEIPT", indicators, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}),
	}
}

func buildView(audience, resolution string, indicators []Indicator, indexes []int) View {
	view := View{Audience: audience, Resolution: resolution, Total: len(indexes)}
	for _, index := range indexes {
		indicator := indicators[index]
		view.IndicatorIDs = append(view.IndicatorIDs, indicator.ID)
		if indicator.Satisfied {
			view.Satisfied++
		}
	}
	view.BasisPoints = coordinate(view.Satisfied, view.Total).BasisPoints
	return view
}

func buildProofs(report Report, checks evidence) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", Claim: "the Gooo source explicitly declares a value program and lower semantic loss fails closed", MetaOperation: "bind-explicit-source", EvidenceDigest: report.SourceDigest, Passed: checks.sourceParsed && checks.programPresent && checks.coreIRFailClosed},
		{Choice: "COHERENCE", Claim: "the activity model, operation registry, and semantic fingerprint agree", MetaOperation: "compile-registry-bound-program", EvidenceDigest: digestValue([]string{report.SemanticFingerprint, report.ValueProgramDigest}), Passed: checks.semanticBound && checks.fingerprintSensitive && checks.registryKnown && checks.operandParsed && checks.signatureSupported},
		{Choice: "REGRESSION", Claim: "fixed value cases and fail-closed counterexamples replay exactly", MetaOperation: "replay-value-witness-corpus", EvidenceDigest: digestValue(struct {
			Cases           []CaseResult
			Counterexamples []CounterexampleResult
		}{report.Cases, report.Counterexamples}), Passed: checks.valueCasesExact && checks.outputsObserved && checks.deterministicReplay && checks.counterexamplesExact && checks.overflowFailClosed},
	}
}

func allIndicatorsSatisfied(indicators []Indicator) bool {
	if len(indicators) != 16 {
		return false
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return false
		}
	}
	return true
}
