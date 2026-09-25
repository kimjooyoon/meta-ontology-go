package guardedpromotion

func Evaluate(source Source, coordinates []Coordinate, summary Summary) (string, string, string) {
	if coordinates[0].Status != statusSatisfied {
		return DecisionFailClosed, ReasonRepositoryMismatch, ResolutionLower
	}
	if source.CollectionError != "" || source.AmbiguousCandidates > 0 || summary.Unresolved > 0 {
		return DecisionFailClosed, ReasonEvidenceUnknown, ResolutionLower
	}
	if summary.ReadinessPromotionAuthorized {
		return DecisionAuthorized, ReasonAuthorized, ResolutionExact
	}
	if onlyMergedBoundaryMissing(coordinates) {
		return DecisionDenied, ReasonMergedPushRequired, ResolutionExact
	}
	return DecisionDenied, ReasonGuardrailRejected, ResolutionExact
}

func onlyMergedBoundaryMissing(coordinates []Coordinate) bool {
	for _, coordinate := range coordinates {
		if coordinate.Status == statusSatisfied {
			continue
		}
		if coordinate.ID != "merged-push-event" &&
			coordinate.ID != "default-branch-boundary" {
			return false
		}
	}
	return true
}
