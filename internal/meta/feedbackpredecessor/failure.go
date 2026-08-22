package feedbackpredecessor

func failureReason(summary Summary) string {
	switch {
	case summary.RepositoryWrites != 0:
		return ReasonWriteEffect
	case summary.ExactHeadCandidates == 0:
		return ReasonNotFound
	case summary.CanonicalCandidates == 0:
		return ReasonCanonicalUnbound
	case summary.SuccessfulCandidates == 0:
		return ReasonUnsuccessful
	case summary.AvailableCandidates == 0:
		return ReasonUnavailable
	case summary.ReceiptBoundCandidates == 0:
		return ReasonReceiptUnbound
	case summary.AmbiguousCandidates != 0:
		return ReasonAmbiguous
	default:
		return ""
	}
}
