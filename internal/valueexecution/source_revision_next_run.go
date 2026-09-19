package valueexecution

const AcceptedRevisionNextRunComparisonSchema = "gooo/value-execution-accepted-revision-next-run/v1"

const (
	NextRunOutcomeImproved = "IMPROVED"
	NextRunOutcomeUnknown  = "UNKNOWN"
	NextRunOutcomeRefuted  = "REFUTED"
)

type AcceptedRevisionNextRunRequest struct {
	Revision          SourceRevision
	Evaluation        SourceRevisionEvaluation
	Accepted          AcceptedSourceRevisionExecution
	BaselineFilename  string
	BaselineSource    []byte
	CandidateFilename string
	CandidateSource   []byte
	Activity          string
	Input             int64
}

type AcceptedRevisionNextRunComparison struct {
	Schema                       string      `json:"schema"`
	State                        ReplayState `json:"state"`
	Outcome                      string      `json:"outcome"`
	Reason                       string      `json:"reason"`
	NextOperation                string      `json:"next_operation"`
	BlockedBy                    []string    `json:"blocked_by"`
	SourceDigest                 string      `json:"source_digest"`
	CandidateSourceDigest        string      `json:"candidate_source_digest"`
	RevisionCandidateID          string      `json:"revision_candidate_id"`
	Activity                     string      `json:"activity"`
	InputDigest                  string      `json:"input_digest"`
	BaselineFailureCode          string      `json:"baseline_failure_code,omitempty"`
	BaselineExecutionDigest      string      `json:"baseline_execution_digest,omitempty"`
	AcceptedExecutionDigest      string      `json:"accepted_execution_digest,omitempty"`
	NextCandidateExecutionDigest string      `json:"next_candidate_execution_digest,omitempty"`
	ExecutionAllowed             bool        `json:"execution_allowed"`
	RepositoryWrites             int         `json:"repository_writes"`
}

// CompareAcceptedRevisionNextRun closes the missing causal step after an
// explicit accepted-candidate reexecution. It repeats the baseline failure and
// candidate success under the same source, candidate, activity, and input
// identities. It records improvement evidence only; it never adopts source or
// grants repository-write authority.
func CompareAcceptedRevisionNextRun(request AcceptedRevisionNextRunRequest) AcceptedRevisionNextRunComparison {
	sourceDigest := digestBytes(request.BaselineSource)
	candidateDigest := digestBytes(request.CandidateSource)
	inputDigest := digestValue(map[string]int64{request.Activity: request.Input})
	comparison := AcceptedRevisionNextRunComparison{
		Schema:        AcceptedRevisionNextRunComparisonSchema,
		State:         ReplayUnknown,
		Outcome:       NextRunOutcomeUnknown,
		Reason:        "NEXT_RUN_SCOPE_UNKNOWN",
		NextOperation: "CAPTURE_MATCHING_NEXT_RUN_INPUTS",
		BlockedBy:     []string{"revision_identity", "source_digests", "activity", "input_digest"},
		SourceDigest:  sourceDigest, CandidateSourceDigest: candidateDigest,
		RevisionCandidateID: request.Revision.CandidateID, Activity: request.Activity,
		InputDigest: inputDigest, ExecutionAllowed: false, RepositoryWrites: 0,
	}
	if request.Revision.Schema != SourceRevisionSchema || request.Revision.CandidateID == "" || request.Revision.ExecutionAllowed || request.Revision.RepositoryWrites != 0 || request.Revision.TriggerReason == "" ||
		request.Evaluation.Schema != SourceRevisionEvaluationSchema || request.Evaluation.State != ReplayClosed || !request.Evaluation.Accepted || !request.Evaluation.CandidateExecuted || request.Evaluation.RepositoryWrites != 0 ||
		request.Accepted.Schema != AcceptedSourceRevisionExecutionSchema || request.Accepted.Decision != "PASS" || request.Accepted.ExecutionAllowed || request.Accepted.RepositoryWrites != 0 ||
		request.Revision.SourceDigest != sourceDigest || request.Revision.CandidateSourceDigest != candidateDigest || request.Revision.Activity != request.Activity ||
		request.Evaluation.SourceDigest != sourceDigest || request.Evaluation.CandidateSourceDigest != candidateDigest || request.Evaluation.Activity != request.Activity || request.Evaluation.InputDigest != inputDigest ||
		request.Accepted.SourceDigest != sourceDigest || request.Accepted.CandidateSourceDigest != candidateDigest || request.Accepted.RevisionCandidateID != request.Revision.CandidateID || request.Accepted.Activity != request.Activity || request.Accepted.InputDigest != inputDigest {
		return comparison
	}
	if !replayReceiptShapeValid(request.Accepted.Execution) || executionDigest(request.Accepted.Execution) != request.Accepted.Execution.ExecutionDigest {
		return nextRunRefuted(comparison, "ACCEPTED_EXECUTION_RECEIPT_INVALID", "PRESERVE_NEXT_RUN_COUNTEREXAMPLE")
	}
	comparison.AcceptedExecutionDigest = request.Accepted.Execution.ExecutionDigest

	baselinePlan, err := CompilePlan(request.BaselineFilename, request.BaselineSource)
	if err != nil {
		return nextRunRefuted(comparison, "SOURCE_REVISION_BASELINE_NOT_EXECUTABLE", "PRESERVE_BASELINE_COMPILE_FAILURE")
	}
	baselineExecution, err := baselinePlan.Execute(map[string]int64{request.Activity: request.Input})
	comparison.BaselineExecutionDigest = baselineExecution.ExecutionDigest
	if err == nil {
		return nextRunRefuted(comparison, "SOURCE_REVISION_BASELINE_NOT_FAILED", "PRESERVE_NON_FAILING_BASELINE")
	}
	failure, ok := FailureOf(err)
	if !ok || failure.Code != request.Revision.TriggerReason {
		return nextRunRefuted(comparison, "SOURCE_REVISION_BASELINE_FAILURE_MISMATCH", "PRESERVE_BASELINE_COUNTEREXAMPLE")
	}
	comparison.BaselineFailureCode = failure.Code

	candidatePlan, err := CompilePlan(request.CandidateFilename, request.CandidateSource)
	if err != nil {
		return nextRunRefuted(comparison, "SOURCE_REVISION_CANDIDATE_NOT_RECOVERED", "PRESERVE_CANDIDATE_COMPILE_FAILURE")
	}
	candidateExecution, err := candidatePlan.Execute(map[string]int64{request.Activity: request.Input})
	comparison.NextCandidateExecutionDigest = candidateExecution.ExecutionDigest
	if err != nil {
		return nextRunRefuted(comparison, "SOURCE_REVISION_CANDIDATE_NOT_RECOVERED", "PRESERVE_CANDIDATE_COUNTEREXAMPLE")
	}
	if candidateExecution.ExecutionDigest != request.Accepted.Execution.ExecutionDigest {
		return nextRunRefuted(comparison, "SOURCE_REVISION_NEXT_RUN_MISMATCH", "PRESERVE_NEXT_RUN_COUNTEREXAMPLE")
	}
	comparison.State = ReplayClosed
	comparison.Outcome = NextRunOutcomeImproved
	comparison.Reason = "SOURCE_REVISION_IMPROVED_ON_NEXT_RUN"
	comparison.NextOperation = "RECORD_IMPROVEMENT_EVIDENCE"
	comparison.BlockedBy = []string{}
	return comparison
}

func nextRunRefuted(comparison AcceptedRevisionNextRunComparison, reason, next string) AcceptedRevisionNextRunComparison {
	comparison.State = ReplayRefuted
	comparison.Outcome = NextRunOutcomeRefuted
	comparison.Reason = reason
	comparison.NextOperation = next
	comparison.BlockedBy = []string{}
	return comparison
}
