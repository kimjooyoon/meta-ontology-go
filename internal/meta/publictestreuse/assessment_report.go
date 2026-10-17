package publictestreuse

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

// VerifyAssessment checks the retained assessment and report for consistency.
// It does not re-execute a policy, authenticate supplied facts or authorize reuse.
func VerifyAssessment(report Report) error {
	value := report.PolicyEvaluation
	if report.Schema != ReportSchema || report.PolicyEvaluations != 1 || value == nil || len(report.PolicyInputs) == 0 ||
		value.SourceSHA256 != report.PolicyDecisionSourceDigest || !knownSHA256Identity(value.SourceSHA256) ||
		!cache.Digest(report.Binding.PolicySourceDigest).Known() ||
		!knownSHA256Identity(value.ProjectionSHA256) {
		return errors.New("report has no source-bound Gooo assessment")
	}
	var activityID string
	switch report.Operation {
	case OriginalOperation:
		activityID = "gooo://test-reuse/assess-original/v1"
		var input OriginalObservation
		if err := decodeAssessmentObject(report.PolicyInputs, &input); err != nil {
			return err
		}
		if report.Decision == DecisionClosed && (!input.BuildPassed || !input.TestPassed || !input.SemanticMatches) {
			return errors.New("completed original report contradicts its execution observations")
		}
	case ReplayOperation:
		activityID = "gooo://test-reuse/assess-replay/v1"
		var input ReuseObservation
		if err := decodeAssessmentObject(report.PolicyInputs, &input); err != nil {
			return err
		}
		if report.ReusedTestExecutions != 0 && !input.Requested {
			return errors.New("retained reuse has no explicit request observation")
		}
	default:
		return errors.New("Gooo assessment has an unknown operation")
	}
	if value.ActivityID != activityID {
		return errors.New("Gooo assessment activity differs from the reported operation")
	}
	var assessment Assessment
	if err := decodeAssessmentObject(value.Value, &assessment); err != nil {
		return err
	}
	if err := assessment.validate(); err != nil {
		return err
	}
	if report.Operation == OriginalOperation && (assessment.Reuse || assessment.CaseID != CaseBaselineExecution) ||
		report.Operation == ReplayOperation && (assessment.Decision == DecisionClosed) != assessment.Reuse {
		return errors.New("Gooo assessment contradicts its reported operation")
	}
	if report.CaseID != assessment.CaseID || report.Decision != assessment.Decision || report.Reason != assessment.Reason ||
		assessment.Reuse != (report.ReusedTestExecutions == 1) || report.ReusedTestExecutions < 0 || report.ReusedTestExecutions > 1 {
		return errors.New("report differs from its retained Gooo assessment")
	}
	var unknown *UnknownState
	if assessment.Decision == DecisionUnknown {
		unknown = Unknown(assessment.Reason, assessment.Step, assessment.NextOperation, assessment.BlockedBy)
	}
	if !SameUnknown(report.Unknown, unknown) {
		return errors.New("report next operation differs from its retained Gooo assessment")
	}
	return nil
}
