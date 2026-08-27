package audienceresolution

import "reflect"

func Evaluate(input Input) Receipt {
	if !ContractValid(input.Contract) {
		return closedReceipt(input, "LOWER_RESOLUTION", "CONTRACT_DENOMINATOR_INVALID")
	}
	state := inspectRecords(input.Ledger)
	sourceBound := sourceMatches(input)
	replay := reflect.DeepEqual(input.Ledger, input.Replay)
	nesting := coordinateNesting(input.Contract)
	resolutions := validResolutions(input.Contract)
	indicators := buildIndicators(input, state, sourceBound, replay, nesting, resolutions,
		counterexamplesValid(input.Ledger.Counterexamples))
	satisfied := countSatisfied(indicators)
	decision, resolution, reason := "PASS", "EXACT", "AUDIENCE_RESOLUTION_OBSERVED"
	if state.missing || !sourceBound {
		decision, resolution, reason = "FAIL_CLOSED", "LOWER_RESOLUTION", "AUDIENCE_EVIDENCE_MISSING"
	} else if state.contradict {
		decision, resolution, reason = "FAIL_CLOSED", "INVARIANT_ONLY", "AUDIENCE_DECISION_CONTRADICTION"
	} else if satisfied != IndicatorTotal {
		decision, resolution, reason = "FAIL_CLOSED", "INVARIANT_ONLY", "AUDIENCE_PROJECTION_CONTRACT_NOT_SATISFIED"
	}
	receipt := Receipt{Schema: ReceiptSchema, ContractID: input.Contract.ID, Subject: input.Ledger.Subject,
		Decision: decision, Resolution: resolution, Reason: reason, Source: input.Ledger.Source,
		Summary: Summary{Coordinates: Coordinates{Satisfied: satisfied, Total: IndicatorTotal,
			BasisPoints: satisfied * 10000 / IndicatorTotal}, RecordsObserved: len(input.Ledger.Records),
			CounterexamplesBlocked: blockedCounterexamples(input.Ledger.Counterexamples),
			MissingCoordinates:     boolInt(state.missing), ContradictoryCoordinates: boolInt(state.contradict)},
		Indicators: indicators, Views: buildViews(input.Contract, indicators, decision, reason),
		Counterexamples:  append([]Counterexample(nil), input.Ledger.Counterexamples...),
		ClaimTransitions: claimTransitions(indicators), NotClaimed: append([]string(nil), input.Contract.NotClaimed...),
		FactsDigest: factsDigest(input.Ledger)}
	return seal(receipt)
}

func closedReceipt(input Input, resolution, reason string) Receipt {
	indicators := make([]Indicator, 0, IndicatorTotal)
	for _, spec := range indicatorSpecs() {
		indicators = append(indicators, Indicator{ID: spec.ID, Class: spec.Class, Producer: spec.Producer,
			Consumer: spec.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
			Stage: spec.Stage, Step: spec.Step, Reason: spec.Reason, ClaimBefore: "UNPROVEN",
			ClaimAfter: "BLOCKED", Expected: 1})
	}
	return seal(Receipt{Schema: ReceiptSchema, ContractID: input.Contract.ID, Subject: input.Ledger.Subject,
		Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason, Source: input.Ledger.Source,
		Summary: Summary{Coordinates: Coordinates{Total: IndicatorTotal, BasisPoints: 0}, RecordsObserved: len(input.Ledger.Records),
			CounterexamplesBlocked: blockedCounterexamples(input.Ledger.Counterexamples)}, Indicators: indicators,
		Views:            buildViews(CanonicalContract(), indicators, "FAIL_CLOSED", reason),
		Counterexamples:  append([]Counterexample(nil), input.Ledger.Counterexamples...),
		ClaimTransitions: claimTransitions(indicators), NotClaimed: append([]string(nil), CanonicalContract().NotClaimed...),
		FactsDigest: factsDigest(input.Ledger)})
}

func sourceMatches(input Input) bool {
	return input.SourcePath == input.Contract.SourcePath && input.Ledger.Source.Path == input.Contract.SourcePath &&
		input.Ledger.Source.Kind == SourceKind && input.Ledger.Source.DeclarationCount == input.Contract.SourceDeclarationCount &&
		input.Ledger.Source.Digest == digestBytes(input.Source) && validDigest(input.Ledger.Source.Digest) &&
		declarationCount(input.Source) == input.Contract.SourceDeclarationCount
}

func declarationCount(source []byte) int {
	count := 0
	for _, line := range splitLines(string(source)) {
		if hasDeclarationPrefix(line, "entity ") || hasDeclarationPrefix(line, "activity ") {
			count++
		}
	}
	return count
}

func hasDeclarationPrefix(line, prefix string) bool {
	line = trimSpace(line)
	return len(line) >= len(prefix) && line[:len(prefix)] == prefix
}

func splitLines(value string) []string {
	result := []string{}
	start := 0
	for index, character := range value {
		if character == '\n' {
			result = append(result, value[start:index])
			start = index + 1
		}
	}
	return append(result, value[start:])
}

func trimSpace(value string) string {
	start, end := 0, len(value)
	for start < end && (value[start] == ' ' || value[start] == '\t' || value[start] == '\r') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\t' || value[end-1] == '\r') {
		end--
	}
	return value[start:end]
}

func countSatisfied(indicators []Indicator) int {
	count := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			count++
		}
	}
	return count
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
