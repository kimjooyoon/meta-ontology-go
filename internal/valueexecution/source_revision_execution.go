package valueexecution

const AcceptedSourceRevisionExecutionSchema = "gooo/value-execution-accepted-source-revision/v1"
const AcceptedSourceRevisionDecision = "ACCEPT"

type AcceptedSourceRevisionRequest struct {
	Revision          SourceRevision
	Evaluation        SourceRevisionEvaluation
	BaselineFilename  string
	BaselineSource    []byte
	CandidateFilename string
	CandidateSource   []byte
	Activity          string
	Input             int64
	ExplicitDecision  string
}

type AcceptedSourceRevisionExecution struct {
	Schema                string              `json:"schema"`
	Decision              string              `json:"decision"`
	Reason                string              `json:"reason"`
	ExplicitDecision      string              `json:"explicit_decision"`
	SourceDigest          string              `json:"source_digest"`
	CandidateSourceDigest string              `json:"candidate_source_digest"`
	RevisionCandidateID   string              `json:"revision_candidate_id"`
	Activity              string              `json:"activity"`
	InputDigest           string              `json:"input_digest"`
	EvaluationState       ReplayState         `json:"evaluation_state"`
	EvaluationReason      string              `json:"evaluation_reason"`
	NextOperation         string              `json:"next_operation"`
	BlockedBy             []string            `json:"blocked_by"`
	ExecutionAllowed      bool                `json:"execution_allowed"`
	AdoptionAuthorized    bool                `json:"adoption_authorized"`
	RepositoryWrites      int                 `json:"repository_writes"`
	Execution             Execution           `json:"execution"`
	AnalysisProvenance    *AnalysisProvenance `json:"analysis_provenance,omitempty"`
}

// ExecuteAcceptedSourceRevision re-executes a candidate only after an
// explicit caller decision and a closed, digest-bound evaluation. It never
// edits source or grants repository adoption authority.
func ExecuteAcceptedSourceRevision(request AcceptedSourceRevisionRequest) (AcceptedSourceRevisionExecution, error) {
	if request.ExplicitDecision != AcceptedSourceRevisionDecision {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "require-explicit-decision", "source revision execution requires an explicit ACCEPT decision")
	}
	if request.Activity == "" || request.BaselineFilename == "" || request.CandidateFilename == "" {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "require-execution-identity", "baseline, candidate, and activity are required")
	}
	if request.Revision.Schema != SourceRevisionSchema || request.Revision.CandidateID == "" || request.Revision.ExecutionAllowed || request.Revision.RepositoryWrites != 0 {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "validate-revision-authority", "source revision is not a non-executing candidate")
	}
	if request.Revision.AnalysisProvenance != nil && !request.Revision.AnalysisProvenance.validFor(digestBytes(request.BaselineSource)) {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "validate-analysis-provenance", "source revision analysis provenance is not source-bound")
	}
	if request.Evaluation.Schema != SourceRevisionEvaluationSchema || request.Evaluation.State != ReplayClosed || !request.Evaluation.Accepted || !request.Evaluation.CandidateExecuted || !request.Evaluation.ContractPreservation || len(request.Evaluation.ContractInputs) == 0 || request.Evaluation.RepositoryWrites != 0 {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "validate-evaluation-closure", "source revision evaluation is not an accepted closed observation")
	}
	if !sameAnalysisProvenance(request.Revision.AnalysisProvenance, request.Evaluation.AnalysisProvenance) {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "bind-analysis-provenance", "revision and evaluation provenance do not agree")
	}
	sourceDigest := digestBytes(request.BaselineSource)
	candidateDigest := digestBytes(request.CandidateSource)
	inputDigest := digestValue(map[string]int64{request.Activity: request.Input})
	if request.Revision.SourceDigest != sourceDigest || request.Revision.CandidateSourceDigest != candidateDigest || request.Revision.Activity != request.Activity ||
		request.Evaluation.SourceDigest != sourceDigest || request.Evaluation.CandidateSourceDigest != candidateDigest || request.Evaluation.Activity != request.Activity || request.Evaluation.InputDigest != inputDigest {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "bind-evaluation-digests", "revision, evaluation, source, candidate, or input identity does not match")
	}
	current := EvaluateSourceRevision(request.Revision, request.BaselineFilename, request.BaselineSource, request.CandidateFilename, request.CandidateSource, request.Activity, request.Input)
	verified := VerifySourceRevisionContract(request.Revision, current, request.BaselineFilename, request.BaselineSource, request.CandidateFilename, request.CandidateSource, request.Activity, SourceRevisionContract{Scope: SourceRevisionContractScope, Inputs: request.Evaluation.ContractInputs, ExpectedOutputs: request.Evaluation.ContractExpectedOutputs})
	if !verified.ContractPreservation || verified.ContractDigest != request.Evaluation.ContractDigest {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "replay-contract-preservation", "source revision contract evidence is not reproducible")
	}
	if !replayReceiptShapeValid(request.Evaluation.CandidateExecution) {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "validate-candidate-receipt", "accepted evaluation has no complete candidate execution receipt")
	}
	plan, err := CompilePlan(request.CandidateFilename, request.CandidateSource)
	if err != nil {
		return AcceptedSourceRevisionExecution{}, err
	}
	execution, err := plan.Execute(map[string]int64{request.Activity: request.Input})
	if err != nil {
		return AcceptedSourceRevisionExecution{}, err
	}
	if execution.ExecutionDigest != request.Evaluation.CandidateExecution.ExecutionDigest {
		return AcceptedSourceRevisionExecution{}, failAt(ReasonSourceRevisionInvalid, "ACCEPT", "compare-candidate-reexecution", "candidate reexecution differs from accepted evaluation")
	}
	return AcceptedSourceRevisionExecution{
		Schema: AcceptedSourceRevisionExecutionSchema, Decision: "PASS", Reason: "SOURCE_REVISION_ACCEPTED_AND_REEXECUTED",
		ExplicitDecision: request.ExplicitDecision, SourceDigest: sourceDigest, CandidateSourceDigest: candidateDigest,
		RevisionCandidateID: request.Revision.CandidateID, Activity: request.Activity, InputDigest: inputDigest,
		EvaluationState: request.Evaluation.State, EvaluationReason: request.Evaluation.Reason,
		NextOperation: "CAPTURE_NEXT_RUN_COMPARISON", BlockedBy: []string{},
		ExecutionAllowed: false, AdoptionAuthorized: true, RepositoryWrites: 0, Execution: execution,
		AnalysisProvenance: cloneAnalysisProvenance(request.Evaluation.AnalysisProvenance),
	}, nil
}

func sameAnalysisProvenance(left, right *AnalysisProvenance) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
