package audienceresolution

func claimTransitions(model semanticSourceModel, recipes []EvidenceRecord, state recordState, sourceDigest string) []ClaimTransition {
	result := make([]ClaimTransition, 0, len(model.Audiences)*len(sourceCoordinates(model)))
	previous := digestBytes([]byte("gooo://audience-resolution/claim-event/genesis"))
	for _, audience := range model.Audiences {
		for _, coordinate := range sourceCoordinates(model) {
			recipe := recipeFor(recipes, coordinate)
			record, visible := state.records[coordinate]
			isVisible := visible && contains(audience.Coordinates, coordinate)
			evidenceStatus := EvidenceUnknown
			evidenceDigest := digestBytes([]byte("unobserved:" + coordinate))
			producer, consumer, metaOperation, proofChoice, stage, step, reason := recipe.Producer, recipe.Consumer, recipe.MetaOperation, recipe.ProofChoice, recipe.Stage, recipe.Step, recipe.Reason
			if record.EvidenceStatus != "" {
				evidenceStatus = record.EvidenceStatus
			}
			if validDigest(record.ContentDigest) {
				evidenceDigest = record.ContentDigest
			}
			if producer == "" {
				producer = "audience-resolution.policy"
			}
			if consumer == "" {
				consumer = audience.Audience
			}
			if metaOperation == "" {
				metaOperation = "project-audience-claim"
			}
			if proofChoice == "" {
				proofChoice = "COHERENCE"
			}
			before := recipe.PriorClaim
			if before == "" {
				before = "OPEN"
			}
			after := "OPEN"
			if isVisible && state.contradict[coordinate] {
				after = "REFUTED"
			} else if isVisible && state.valid[coordinate] {
				after = "DISCHARGED"
			}
			if !isVisible {
				evidenceStatus = EvidenceUnknown
			}
			event := ClaimTransition{ClaimID: audienceClaimID(recipe, audience.Audience), Audience: audience.Audience,
				IndicatorID: coordinate, Proposition: recipe.Proposition, PropositionDigest: recipe.PropositionDigest,
				TargetAddress: recipe.TargetAddress, Before: before, After: after,
				Visibility: visibilityLabel(isVisible), EvidenceStatus: evidenceStatus, EvidenceDigest: evidenceDigest,
				PreviousEventDigest: previous, SourceDigest: sourceDigest, SemanticSourceDigest: model.SemanticDigest,
				Producer: producer, Consumer: consumer, MetaOperation: metaOperation, ProofChoice: proofChoice,
				Stage: stage, Step: step, Reason: reason}
			event.EventDigest = claimEventDigest(event)
			previous = event.EventDigest
			result = append(result, event)
		}
	}
	return result
}

func audienceClaimID(recipe EvidenceRecord, audience string) string {
	claimID := recipe.ClaimID
	if claimID == "" {
		claimID = "claim/" + recipe.Coordinate
	}
	return claimID + "/audience/" + audience
}

func visibilityLabel(visible bool) string {
	if visible {
		return "VISIBLE"
	}
	return "OMITTED"
}

func claimEventDigest(event ClaimTransition) string {
	event.EventDigest = ""
	return digestJSON(event)
}

func finalClaimEventDigest(events []ClaimTransition) string {
	if len(events) == 0 {
		return digestBytes([]byte("gooo://audience-resolution/claim-event/genesis"))
	}
	return events[len(events)-1].EventDigest
}
