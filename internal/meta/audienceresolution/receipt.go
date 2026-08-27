package audienceresolution

func claimTransitions(model semanticSourceModel, state recordState, ledger Ledger) []ClaimTransition {
	result := make([]ClaimTransition, 0, len(model.Audiences)*len(sourceCoordinates(model)))
	for _, audience := range model.Audiences {
		for _, coordinate := range sourceCoordinates(model) {
			before, after := "OPEN", "OPEN"
			visibility := "OMITTED"
			evidenceDigest := digestBytes([]byte("omitted:" + coordinate))
			producer, consumer, stage, step, reason := "audience-resolution.policy", audience.Audience, "projection", "policy", "formal policy coordinate is outside this audience's view"
			if contains(audience.Coordinates, coordinate) {
				if record, ok := state.records[coordinate]; ok {
					if record.PriorClaim != "" {
						before = record.PriorClaim
					}
					visibility = "VISIBLE"
					evidenceDigest = digestBytes([]byte(recordEvidenceCanonical(record)))
					producer, consumer, stage, step, reason = record.Producer, record.Consumer, record.Stage, record.Step, record.Reason
					if state.contradict[coordinate] {
						after = "REFUTED"
					} else if state.valid[coordinate] {
						after = "DISCHARGED"
					}
				}
			}
			result = append(result, ClaimTransition{IndicatorID: coordinate, Audience: audience.Audience, Before: before, After: after,
				Visibility: visibility, EvidenceDigest: evidenceDigest, Producer: producer, Consumer: consumer,
				Stage: stage, Step: step, Reason: reason})
		}
	}
	_ = ledger
	return result
}

func recordEvidenceCanonical(record EvidenceRecord) string {
	return record.ID + "\x00" + record.Coordinate + "\x00" + record.Audience + "\x00" + record.Stage + "\x00" + record.Step + "\x00" + record.Reason + "\x00" + record.Producer + "\x00" + record.Consumer + "\x00" + record.MetaOperation + "\x00" + record.ProofChoice + "\x00" + record.PriorClaim + "\x00" + record.Observation
}

func executeCounterexamplesFromLedger(input Input, model semanticSourceModel) []CounterexampleResult {
	results := make([]CounterexampleResult, 0, len(input.Ledger.Counterexamples))
	baseline := evaluateInput(input, false)
	for _, counterexample := range input.Ledger.Counterexamples {
		mutated := input
		mutated.Ledger = mutateLedger(input.Ledger, counterexample)
		mutated.Replay = mutated.Ledger
		observed := evaluateInput(mutated, false)
		views := make([]CounterexampleView, 0, len(observed.Views))
		for index, view := range observed.Views {
			before := "UNKNOWN"
			if index < len(baseline.Views) {
				before = baseline.Views[index].LocalDecision
			}
			after := view.LocalDecision
			views = append(views, CounterexampleView{Audience: view.Audience, Before: before, After: after,
				LocalDecision: view.LocalDecision, LocalResolution: view.LocalResolution})
		}
		results = append(results, CounterexampleResult{ID: counterexample.ID, Kind: counterexample.Kind,
			Trigger: counterexample.Trigger, Mutation: counterexample.Mutation, Global: observed.Decision,
			Views: views, Transition: counterexample.TargetCoordinate + ":OPEN->" + counterexampleTransition(observed, counterexample.TargetCoordinate), Reason: observed.Reason})
	}
	_ = model
	return results
}

func mutateLedger(ledger Ledger, counterexample Counterexample) Ledger {
	mutated := ledger
	mutated.Records = append([]EvidenceRecord(nil), ledger.Records...)
	if counterexample.Kind == "INFORMATION_OMISSION" {
		filtered := mutated.Records[:0]
		for _, record := range mutated.Records {
			if record.Coordinate != counterexample.TargetCoordinate {
				filtered = append(filtered, record)
			}
		}
		mutated.Records = filtered
		return mutated
	}
	for index := range mutated.Records {
		if mutated.Records[index].Coordinate == counterexample.TargetCoordinate {
			mutated.Records[index].Observation = "CONTRADICTORY"
		}
	}
	return mutated
}

func counterexampleTransition(receipt Receipt, coordinate string) string {
	for _, transition := range receipt.ClaimTransitions {
		if transition.Audience == "GOVERNOR" && transition.IndicatorID == coordinate {
			return transition.After
		}
	}
	return "OPEN"
}
