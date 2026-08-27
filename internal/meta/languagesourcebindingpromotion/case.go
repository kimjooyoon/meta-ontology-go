package languagesourcebindingpromotion

func evaluateCase(definition CaseDefinition, producer, receipt, oracle []byte, head, policyDigest string) CaseResult {
	structural := assessStructural(producer, receipt, head)
	independent := assessIndependent(oracle, receipt, head, structural.ArtifactDigest)
	promotion := promotionClaim(structural, independent, policyDigest)
	decision, resolution, reason, coordinate := DecisionPass, ResolutionExact, "SOURCE_BINDING_CLAIM_DISCHARGED", promotion.Coordinate
	if structural.Status == "REFUTED" {
		decision, resolution, reason, coordinate = DecisionClosed, ResolutionInvariant, structural.Reason, structural.Coordinate
	} else if independent.Status == "REFUTED" {
		decision, resolution, reason, coordinate = DecisionClosed, ResolutionInvariant, independent.Reason, independent.Coordinate
	} else if structural.Status == "OPEN" {
		decision, resolution, reason, coordinate = DecisionClosed, ResolutionLower, structural.Reason, structural.Coordinate
	} else if independent.Status == "OPEN" {
		decision, resolution, reason, coordinate = DecisionClosed, ResolutionLower, independent.Reason, independent.Coordinate
	}
	result := CaseResult{ID: definition.ID, ExpectedDecision: definition.ExpectedDecision,
		ObservedDecision: decision, ObservedResolution: resolution, ObservedReason: reason,
		Coordinate: coordinate, Claims: []ClaimResult{
			claim("structural-source-execution", structural), claim("independent-source-binding", independent),
			claim("source-binding-promotion", promotion)},
		Status: "NOT_SATISFIED"}
	if decision == definition.ExpectedDecision && resolution == definition.ExpectedResolution &&
		reason == definition.ExpectedReason && promotion.Status == definition.ExpectedPromotionStatus {
		result.Status = "SATISFIED"
	}
	return result
}

func promotionClaim(structural, independent component, policyDigest string) component {
	evidence := append(append([]string{}, structural.Evidence...), independent.Evidence...)
	evidence = append(evidence, policyDigest)
	if structural.Status == "REFUTED" {
		return refuted(structural.Reason, "PROMOTION_DEPENDENCY", "structural-source-execution", evidence...)
	}
	if independent.Status == "REFUTED" {
		return refuted(independent.Reason, "PROMOTION_DEPENDENCY", "independent-source-binding", evidence...)
	}
	if structural.Status == "OPEN" || independent.Status == "OPEN" {
		value := open("SOURCE_BINDING_PROMOTION_DEPENDENCY_BLOCKED", "PROMOTION_DEPENDENCY", "resolve-claims", evidence...)
		value.UnknownClass = "DEPENDENCY_BLOCKED"
		return value
	}
	return discharged("SOURCE_BINDING_CLAIM_DISCHARGED", "PROMOTION_APPLY", "claim-status", evidence...)
}
