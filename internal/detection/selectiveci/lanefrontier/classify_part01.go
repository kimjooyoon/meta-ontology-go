package lanefrontier

// Classify evaluates facts in the documented fixed order and returns a sealed
// canonical result. It has no side effects and does not select a priority.
func Classify(input Input) Output {
	result := outputFromInput(input)
	if input.SchemaVersion != SchemaVersion {
		return seal(result, DecisionUnknown, ReasonUnknownSchema)
	}
	if !hasRequiredFacts(input) {
		return seal(result, DecisionUnknown, ReasonMissingInput)
	}
	if hasInvalidCount(input) {
		return seal(result, DecisionUnknown, ReasonInvalidCount)
	}

	owners, ownerErr := normalizeOwners(input.OwnedPathPrefixes)
	if ownerErr != nil {
		if _, invalid := ownerErr.(errInvalidPath); invalid {
			return seal(result, DecisionUnknown, ReasonMissingInput)
		}
		return seal(result, DecisionUnknown, ReasonAmbiguousOwner)
	}
	paths, pathErr := normalizePaths(input.ChangedPaths)
	if pathErr != nil {
		return seal(result, DecisionUnknown, ReasonMissingInput)
	}
	if !pathsInScope(paths, owners) {
		return seal(result, DecisionIneligible, ReasonPathOutOfScope)
	}
	result.OwnedPathPrefixes = owners
	result.ChangedPaths = paths

	if input.ActiveLeaseCount > 0 {
		return seal(result, DecisionIneligible, ReasonActiveLease)
	}
	if input.OpenPRCount > 0 {
		return seal(result, DecisionIneligible, ReasonActivePR)
	}
	if input.AheadCount > 0 && input.BehindCount > 0 {
		return seal(result, DecisionIneligible, ReasonDivergedBranch)
	}
	if input.AheadCount == 0 && input.BehindCount > 0 {
		return seal(result, DecisionIneligible, ReasonStaleBranch)
	}
	if input.AheadCount > 0 && input.BehindCount == 0 {
		return seal(result, DecisionIneligible, ReasonBranchAhead)
	}
	return seal(result, DecisionEligible, ReasonEligible)
}

// Evaluate is a descriptive alias for callers that treat the classifier as a
// predicate evaluator.
func Evaluate(input Input) Output { return Classify(input) }
func outputFromInput(input Input) Output {
	return Output{
		SchemaVersion:     input.SchemaVersion,
		RegistryDigest:    input.RegistryDigest,
		BaseSHA:           input.BaseSHA,
		LaneHeadSHA:       input.LaneHeadSHA,
		LaneID:            input.LaneID,
		RegisteredBranch:  input.RegisteredBranch,
		OwnedPathPrefixes: append([]string(nil), input.OwnedPathPrefixes...),
		ChangedPaths:      append([]string(nil), input.ChangedPaths...),
		AheadCount:        input.AheadCount,
		BehindCount:       input.BehindCount,
		OpenPRCount:       input.OpenPRCount,
		ActiveLeaseCount:  input.ActiveLeaseCount,
	}
}
