package artifactfeedback

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"

func EvaluateWithResolution(input ResolutionInput) (ResolutionReport, error) {
	feedback, err := Evaluate(input.Feedback)
	report := ResolutionReport{
		Schema: ResolutionFeedbackSchema, Feedback: feedback,
		SourceDecision: input.Feedback.Coverage.Decision,
		FromResolution: input.CurrentResolution, ToResolution: input.CurrentResolution,
		PreviousDescents: input.Descents, Descents: input.Descents,
		RepositoryWrites: input.Feedback.RepositoryWrites,
	}
	if err != nil {
		return report, err
	}
	conflict := isSemanticCoverageConflict(feedback, report.SourceDecision)
	if !conflict {
		report.Decision = feedback.Decision
		report.Reason = feedback.Reason
		report.NextOperation = feedback.NextOperation
		return finalizeResolutionReport(report, false), nil
	}
	transition := semanticresolution.ResolveSemanticConflict(
		semanticresolution.Conflict{
			SourceDecision:    input.Feedback.Coverage.Decision,
			CurrentResolution: input.CurrentResolution,
			Descents:          input.Descents,
			RepositoryWrites:  input.Feedback.RepositoryWrites,
		},
	)
	report.Decision = transition.Decision
	report.Reason = transition.Reason
	report.Descents = transition.Descents
	if transition.ToResolution != "" {
		report.ToResolution = transition.ToResolution
	}
	if transition.Decision == DecisionLowerResolution {
		report.NextOperation = NextOperationReevaluateFeedback
	}
	return finalizeResolutionReport(report, true), nil
}

func finalizeResolutionReport(report ResolutionReport, conflict bool) ResolutionReport {
	report.Indicators = resolutionIndicators(report, conflict)
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}

func isSemanticCoverageConflict(feedback Report, sourceDecision string) bool {
	if sourceDecision == "" || sourceDecision == "FIXED_POINT" ||
		sourceDecision == "IMPROVE" || feedback.Decision != "FAIL_CLOSED" {
		return false
	}
	if feedback.Reason == ReasonCoverageDecisionUnknown {
		return true
	}
	summary := feedback.Summary
	return feedback.Reason == "FEEDBACK_INPUT_UNBOUND" &&
		summary.RequiredInputs > 0 &&
		summary.ExactHeadInputs == summary.RequiredInputs &&
		summary.BoundInputs == summary.RequiredInputs-1 &&
		summary.ReplayBoundInputs == summary.RequiredInputs &&
		summary.StaleInputs == 0 &&
		summary.AmbiguousNextOperations == 0 &&
		summary.RepositoryWrites == 0
}
