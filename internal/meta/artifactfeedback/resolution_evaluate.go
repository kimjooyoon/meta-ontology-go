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
	conflict := feedback.Decision == "FAIL_CLOSED" &&
		feedback.Reason == ReasonCoverageDecisionUnknown
	if !conflict {
		report.Decision = feedback.Decision
		report.Reason = feedback.Reason
		report.NextOperation = feedback.NextOperation
		return finalizeResolutionReport(report, false), nil
	}
	transition := semanticresolution.ResolveSemanticConflict(
		semanticresolution.Conflict{
			SourceDecision: input.Feedback.Coverage.Decision,
			CurrentResolution: input.CurrentResolution,
			Descents: input.Descents,
			RepositoryWrites: input.Feedback.RepositoryWrites,
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
