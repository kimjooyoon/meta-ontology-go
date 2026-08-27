package audienceresolution

// Evaluate consumes only provider-produced evidence. The optional bundle is
// useful to the CI producer boundary; when absent, the package creates an
// in-memory provider bundle for compatibility with package-level callers.
func Evaluate(input Input, provided ...CurrentEvidenceBundle) Receipt {
	model, err := deriveSemanticSource(input.SourcePath, input.Source)
	if err != nil {
		return closedReceipt(input, "LOWER_RESOLUTION", "SOURCE_RECONSTRUCTION_UNAVAILABLE")
	}
	var bundle CurrentEvidenceBundle
	if len(provided) > 0 {
		bundle = provided[0]
	} else {
		bundle, err = ProvideCurrentEvidence(input)
		if err != nil {
			return closedReceipt(input, "LOWER_RESOLUTION", "CURRENT_EVIDENCE_UNAVAILABLE")
		}
	}
	state := inspectCurrentEvidence(input.Ledger.Records, bundle.Records)
	decision, resolution := subjectDecisionFromState(state, model)
	indicators := buildIndicators(input.Ledger.Records, bundle.Records, state, model, bundle.Replay, bundle.Counterexamples, sourceBound(input, model), sourcePolicyValid(model) && sourceAudienceResolutionValid(model), decision)
	views := buildViews(model, input.Ledger.Records, state, decision, resolution)
	subjectIDs := subjectCoordinates(model)
	subjectSatisfied := 0
	for coordinate := range subjectIDs {
		if state.valid[coordinate] {
			subjectSatisfied++
		}
	}
	sealState := EvidenceUnknown
	sealAfter := "OPEN"
	if seal := recordMap(bundle.Records)["receipt.seal"]; seal.EvidenceStatus != "" {
		sealState = seal.EvidenceStatus
	}
	return seal(Receipt{Schema: ReceiptSchema, ContractID: input.Contract.ID, Subject: input.Ledger.Subject,
		Decision: decision, Resolution: resolution, Reason: decisionReason(decision, resolution, state, model), Provisional: true,
		Source: sourceReport(input.Ledger.Source, model), Summary: Summary{
			Coordinates:          Coordinates{Satisfied: subjectSatisfied, Total: len(subjectIDs), BasisPoints: basisPoints(subjectSatisfied, len(subjectIDs))},
			DistinctPropositions: len(sourceCoordinates(model)), RecordsObserved: len(bundle.Records),
			CounterexamplesExecuted: len(bundle.Counterexamples), MissingCoordinates: missingCount(model, state),
			ContradictoryCoordinates: contradictoryCount(state), SourceDenominator: model.DeclarationCount,
			EvidenceCounts: evidenceCounts(input.Ledger.Records, bundle.Records),
			Conformance:    ConformanceSummary{Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION", SealClaimBefore: "OPEN", SealClaimAfter: sealAfter, SealEvidenceStatus: sealState}},
		Indicators: indicators, CurrentEvidence: bundle.Records, Views: views, Replay: bundle.Replay,
		Counterexamples: bundle.Counterexamples, ClaimTransitions: claimTransitions(model, input.Ledger.Records, state, input.Ledger.Source.Digest),
		NotClaimed: append([]string(nil), input.Contract.NotClaimed...), FactsDigest: factsDigest(input.Ledger)})
}

func sourceBound(input Input, model semanticSourceModel) bool {
	return input.Ledger.Source.Path == input.Contract.SourcePath && input.Ledger.Source.Kind == SourceKind &&
		input.Ledger.Source.Digest == digestBytes(input.Source) && validDigest(input.Ledger.Source.Digest) &&
		(model.SemanticDigest != "")
}

func subjectDecisionFromState(state recordState, model semanticSourceModel) (string, string) {
	for coordinate := range subjectCoordinates(model) {
		if state.contradict[coordinate] {
			return "REFUTED", "INVARIANT_ONLY"
		}
	}
	for coordinate := range subjectCoordinates(model) {
		if !state.valid[coordinate] {
			return "UNKNOWN", "LOWER_RESOLUTION"
		}
	}
	return "PASS", "EXACT"
}

func decisionReason(decision, resolution string, state recordState, model semanticSourceModel) string {
	if decision == "REFUTED" {
		for coordinate := range subjectCoordinates(model) {
			if state.contradict[coordinate] {
				return "VISIBLE_EVIDENCE_CONTRADICTION:" + coordinate
			}
		}
		return "VISIBLE_EVIDENCE_CONTRADICTION"
	}
	if decision == "UNKNOWN" {
		return firstMissingReason(model, state)
	}
	if resolution == "EXACT" {
		return "CURRENT_EVIDENCE_SUBJECT_RECONSTRUCTED;SEAL_CONFORMANCE_POSTCONDITION"
	}
	return "CURRENT_EVIDENCE_SUBJECT_UNRESOLVED"
}

func firstMissingReason(model semanticSourceModel, state recordState) string {
	for _, coordinate := range sourceCoordinates(model) {
		if !subjectCoordinates(model)[coordinate] {
			continue
		}
		if state.contradict[coordinate] {
			return "VISIBLE_EVIDENCE_CONTRADICTION:" + coordinate
		}
		if !state.valid[coordinate] {
			if record, ok := state.records[coordinate]; ok {
				return "VISIBLE_EVIDENCE_INSUFFICIENT:" + record.Stage + ":" + record.Step + ":" + record.Reason
			}
			return "REQUIRED_EVIDENCE_OMITTED:projection:policy:" + coordinate
		}
	}
	return "REQUIRED_EVIDENCE_UNAVAILABLE"
}

func missingCount(model semanticSourceModel, state recordState) int {
	count := 0
	for _, coordinate := range sourceCoordinates(model) {
		if !state.valid[coordinate] {
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

func evidenceCounts(recipes, current []EvidenceRecord) EvidenceCounts {
	counts := EvidenceCounts{}
	for _, record := range recipes {
		if record.EvidenceStatus == EvidenceHistorical {
			counts.Historical++
		}
	}
	for _, record := range current {
		switch record.EvidenceStatus {
		case EvidenceCurrent:
			counts.Current++
		case EvidenceHistorical:
			counts.Historical++
		default:
			counts.Unknown++
		}
	}
	return counts
}

func closedReceipt(input Input, resolution, reason string) Receipt {
	return seal(Receipt{Schema: ReceiptSchema, ContractID: input.Contract.ID, Subject: input.Ledger.Subject,
		Decision: "UNKNOWN", Resolution: resolution, Reason: reason, Provisional: true, Source: input.Ledger.Source,
		Summary: Summary{DistinctPropositions: 0, RecordsObserved: 0, SourceDenominator: 0,
			Coordinates: Coordinates{Total: 0}, EvidenceCounts: EvidenceCounts{Unknown: 1},
			Conformance: ConformanceSummary{Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION", SealClaimBefore: "OPEN", SealClaimAfter: "OPEN", SealEvidenceStatus: EvidenceUnknown}},
		NotClaimed: append([]string(nil), input.Contract.NotClaimed...), FactsDigest: factsDigest(input.Ledger)})
}
