package feedbackpredecessor

func decide(report *Report, eligible []Candidate) {
	report.Decision, report.Reason = DecisionFailClosed, failureReason(report.Summary)
	report.Resolution, report.NextOperation = failureResolution(report.Reason), OperationReevaluate
	if report.Reason == ReasonAmbiguous || report.Reason == ReasonWriteEffect {
		report.NextOperation = OperationHalt
	}
	if report.Reason == ReasonUnsuccessful {
		report.Decision = DecisionLower
		return
	}
	if report.Reason != "" {
		return
	}
	candidate := eligible[0]
	report.Decision, report.Reason = DecisionSelected, ReasonSelected
	report.Resolution, report.NextOperation, report.PromotionAuthorized = ResolutionExact, OperationConsume, true
	report.Selected = &Selection{ArtifactID: candidate.ArtifactID, RunID: candidate.RunID,
		RunAttempt: candidate.RunAttempt, ReceiptDigest: candidate.ReceiptDigest}
}

func failureResolution(reason string) string {
	if reason == ReasonAmbiguous || reason == ReasonWriteEffect {
		return ResolutionInvariant
	}
	return ResolutionClass
}

func Consumable(report Report) bool {
	exact := report.Decision == DecisionSelected && report.Reason == ReasonSelected &&
		report.Resolution == ResolutionExact && report.NextOperation == OperationConsume &&
		report.PromotionAuthorized && report.Selected != nil
	lower := report.Decision == DecisionLower && report.Reason == ReasonUnsuccessful &&
		report.Resolution == ResolutionClass && report.NextOperation == OperationReevaluate &&
		!report.PromotionAuthorized && report.Selected == nil
	foundation := report.Decision == DecisionFoundation && report.Reason == ReasonFoundationRegression &&
		report.Foundation != nil && !report.PromotionAuthorized && report.Selected == nil
	return exact || lower || foundation
}
