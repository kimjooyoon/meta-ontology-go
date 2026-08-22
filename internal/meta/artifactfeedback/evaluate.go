package artifactfeedback

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactcoverage"
)

func Evaluate(input Input) (Report, error) {
	program := CanonicalProgram()
	if err := Validate(program); err != nil {
		return Report{}, err
	}
	if err := validateInput(input); err != nil {
		return Report{}, err
	}
	summary := summarize(input)
	decision, reason, next := feedbackDecision(input.Coverage.Decision, input.Coverage.SelectedOperation, summary)
	report := Report{
		Schema: ReportSchema, CommitSHA: input.Coverage.CommitSHA, Repository: input.Coverage.Repository,
		Decision: decision, Reason: reason, NextOperation: next,
		CoverageReportDigest: input.Coverage.ReportDigest, CycleEnvelopeDigest: input.Cycle.EnvelopeDigest,
		ProgramDigest: digestJSON(program), InputDigest: digestJSON(input), Summary: summary,
	}
	report.Indicators = evaluateKPIs(program.Indicators, summary)
	report.Proofs = feedbackProofs(summary)
	report.ReportDigest = digestJSON(report)
	return report, nil
}

func validateInput(input Input) error {
	coverage := input.Coverage
	if coverage.Schema != artifactcoverage.ReportSchema || len(coverage.CommitSHA) != 40 ||
		coverage.Repository == "" || !validCoverageReportDigest(coverage) {
		return fmt.Errorf("coverage feedback identity is malformed")
	}
	cycle := input.Cycle
	if cycle.Schema != CycleSchema || len(cycle.HeadSHA) != 40 || cycle.Status == "" ||
		cycle.CIConclusion == "" || !validBareDigest(cycle.EnvelopeDigest) ||
		!validBareDigest(cycle.ReplayDigest) || !validPrefixedDigest(input.CoverageReplayDigest) ||
		input.RepositoryWrites < 0 || coverage.Summary.RepositoryWrites < 0 {
		return fmt.Errorf("cycle feedback identity is malformed")
	}
	return nil
}

func feedbackDecision(coverageDecision, selected string, summary Summary) (string, string, string) {
	switch {
	case summary.StaleInputs != 0:
		return "FAIL_CLOSED", "FEEDBACK_INPUT_STALE", ""
	case summary.RepositoryWrites != 0:
		return "FAIL_CLOSED", "FEEDBACK_WRITE_EFFECT", ""
	case summary.BoundInputs != summary.RequiredInputs:
		return "FAIL_CLOSED", "FEEDBACK_INPUT_UNBOUND", ""
	case summary.ReplayBoundInputs != summary.RequiredInputs:
		return "FAIL_CLOSED", "FEEDBACK_REPLAY_UNBOUND", ""
	case summary.AmbiguousNextOperations != 0:
		return "FAIL_CLOSED", "FEEDBACK_NEXT_OPERATION_AMBIGUOUS", ""
	case coverageDecision == "IMPROVE":
		return "IMPROVE", "NEXT_META_OPERATION_SELECTED", selected
	case coverageDecision == "FIXED_POINT":
		return "FIXED_POINT", "NEXT_CYCLE_FEEDBACK_FIXED_POINT", ""
	default:
		return "FAIL_CLOSED", "FEEDBACK_COVERAGE_DECISION_UNKNOWN", ""
	}
}
