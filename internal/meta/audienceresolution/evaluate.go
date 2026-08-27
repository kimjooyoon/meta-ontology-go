package audienceresolution

import (
	"reflect"
)

func Evaluate(input Input) Receipt {
	return evaluateInput(input, true)
}

func evaluateInput(input Input, executeCounterexamples bool) Receipt {
	if !ContractValid(input.Contract) {
		return closedReceipt(input, "INVARIANT_ONLY", "CONTRACT_SHAPE_INVALID")
	}
	model, err := deriveSemanticSource(input.SourcePath, input.Source)
	if err != nil {
		return closedReceipt(input, "LOWER_RESOLUTION", "SOURCE_RECONSTRUCTION_UNAVAILABLE")
	}
	state := inspectRecords(input.Ledger)
	sourceBound := sourceMatches(input, model)
	replay := reflect.DeepEqual(input.Ledger, input.Replay)
	policyValid := sourcePolicyValid(model) && sourceAudienceResolutionValid(model)
	globalDecision, globalResolution, globalReason := globalState(input, model, state, sourceBound, replay, policyValid)
	indicators := buildIndicators(input, model, state, sourceBound, replay, policyValid, globalDecision)
	views := buildViews(model, state, globalDecision, globalResolution, globalReason)
	receipt := Receipt{Schema: ReceiptSchema, ContractID: input.Contract.ID, Subject: input.Ledger.Subject,
		Decision: globalDecision, Resolution: globalResolution, Reason: globalReason,
		Source: sourceReport(input.Ledger.Source, model),
		Summary: Summary{Coordinates: Coordinates{Satisfied: countSatisfied(indicators), Total: IndicatorTotal,
			BasisPoints: basisPoints(countSatisfied(indicators), IndicatorTotal)}, RecordsObserved: len(input.Ledger.Records),
			CounterexamplesExecuted: len(input.Ledger.Counterexamples), MissingCoordinates: missingCount(model, state),
			ContradictoryCoordinates: contradictoryCount(state), SourceDenominator: model.DeclarationCount},
		Indicators: indicators, Views: views, ClaimTransitions: claimTransitions(model, state, input.Ledger),
		NotClaimed: append([]string(nil), input.Contract.NotClaimed...), FactsDigest: factsDigest(input.Ledger)}
	if executeCounterexamples {
		receipt.Counterexamples = executeCounterexamplesFromLedger(input, model)
	}
	return seal(receipt)
}

func globalState(input Input, model semanticSourceModel, state recordState, sourceBound, replay, policyValid bool) (string, string, string) {
	if state.duplicate || contradictoryCount(state) > 0 {
		return "REFUTED", "INVARIANT_ONLY", "VISIBLE_EVIDENCE_CONTRADICTION:ledger:observation:duplicate-or-contradictory-record"
	}
	if missingCount(model, state) > 0 {
		return "UNKNOWN", "LOWER_RESOLUTION", firstMissingReason(model, state)
	}
	if !sourceBound {
		return "UNKNOWN", "LOWER_RESOLUTION", "SOURCE_RECONSTRUCTION_MISMATCH"
	}
	if !replay {
		return "UNKNOWN", "INVARIANT_ONLY", "LEDGER_REPLAY_DIVERGED"
	}
	if !policyValid {
		return "UNKNOWN", "LOWER_RESOLUTION", "SOURCE_POLICY_INVALID"
	}
	if !counterexamplesValid(input.Ledger.Counterexamples) {
		return "UNKNOWN", "LOWER_RESOLUTION", "COUNTEREXAMPLE_DEFINITION_INCOMPLETE"
	}
	return "PASS", "EXACT", "CANONICAL_EVIDENCE_LEDGER_RECONSTRUCTED"
}

func sourceMatches(input Input, model semanticSourceModel) bool {
	// Source.Path is the ledger's logical binding. The bytes supplied at
	// SourcePath may be a CI intervention variant, so their raw digest—not a
	// temporary filename—establishes identity.
	return input.Ledger.Source.Path == input.Contract.SourcePath &&
		input.Ledger.Source.Kind == SourceKind && input.Ledger.Source.Digest == digestBytes(input.Source) &&
		(input.Ledger.Source.SemanticDigest == "" || input.Ledger.Source.SemanticDigest == model.SemanticDigest) && validDigest(input.Ledger.Source.Digest)
}

func sourceReport(raw SourceBinding, model semanticSourceModel) SourceBinding {
	raw.DeclarationCount = model.DeclarationCount
	raw.SemanticDigest = model.SemanticDigest
	raw.Reconstructed = model.CanonicalIRDigest == model.SemanticDigest
	return raw
}

func missingCount(model semanticSourceModel, state recordState) int {
	count := 0
	for _, coordinate := range allCoordinates(model) {
		if !coordinateVisible(state, coordinate) || !state.valid[coordinate] {
			count++
		}
	}
	return count
}

func contradictoryCount(state recordState) int {
	count := 0
	for _, value := range state.contradict {
		if value {
			count++
		}
	}
	if state.duplicate {
		count++
	}
	return count
}

func firstMissingReason(model semanticSourceModel, state recordState) string {
	for _, coordinate := range allCoordinates(model) {
		if coordinateContradictory(state, coordinate) {
			return "VISIBLE_EVIDENCE_CONTRADICTION:" + coordinate
		}
		if !coordinateVisible(state, coordinate) || !state.valid[coordinate] {
			for _, spec := range indicatorSpecs() {
				if spec.ID == coordinate {
					return "REQUIRED_EVIDENCE_OMITTED:" + spec.Stage + ":" + spec.Step + ":" + spec.Reason
				}
			}
			return "REQUIRED_EVIDENCE_OMITTED:projection:policy:" + coordinate
		}
	}
	if !sourceAudienceResolutionValid(model) {
		return "SOURCE_POLICY_RESOLUTION_INVALID"
	}
	if !sourcePolicyValid(model) {
		return "SOURCE_POLICY_NESTING_INVALID"
	}
	return "REQUIRED_EVIDENCE_UNAVAILABLE"
}

func closedReceipt(input Input, resolution, reason string) Receipt {
	model := semanticSourceModel{Audiences: CanonicalContract().Audiences, ClaimStates: []string{"OPEN", "DISCHARGED", "REFUTED"}, PriorClaim: "OPEN", EvidenceClaimRelation: "unavailable"}
	indicators := make([]Indicator, 0, IndicatorTotal)
	for _, spec := range indicatorSpecs() {
		indicators = append(indicators, Indicator{ID: spec.ID, Class: spec.Class, Producer: spec.Producer,
			Consumer: spec.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
			Stage: spec.Stage, Step: spec.Step, Reason: spec.Reason, ClaimBefore: "OPEN", ClaimAfter: "OPEN", Expected: 1})
	}
	return seal(Receipt{Schema: ReceiptSchema, ContractID: input.Contract.ID, Subject: input.Ledger.Subject,
		Decision: "UNKNOWN", Resolution: resolution, Reason: reason, Source: input.Ledger.Source,
		Summary: Summary{Coordinates: Coordinates{Total: IndicatorTotal}, RecordsObserved: len(input.Ledger.Records),
			CounterexamplesExecuted: 0}, Indicators: indicators,
		Views:      buildViews(model, inspectRecords(input.Ledger), "UNKNOWN", resolution, reason),
		NotClaimed: append([]string(nil), CanonicalContract().NotClaimed...), FactsDigest: factsDigest(input.Ledger)})
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
