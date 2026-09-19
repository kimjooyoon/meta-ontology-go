package valueexecution

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const SourceRevisionSchema = "gooo/value-execution-source-revision/v1"
const SourceRevisionEvaluationSchema = "gooo/value-execution-source-revision-evaluation/v1"

type SourceRevisionRequest struct {
	SourceDigest      string `json:"source_digest"`
	Activity          string `json:"activity"`
	ExpectedProgram   string `json:"expected_program"`
	ReplacementProgram string `json:"replacement_program"`
	TriggerReason     string `json:"trigger_reason"`
}

type SourceRevision struct {
	Schema                string `json:"schema"`
	CandidateID           string `json:"candidate_id"`
	SourceDigest          string `json:"source_digest"`
	CandidateSourceDigest string `json:"candidate_source_digest"`
	Activity              string `json:"activity"`
	BeforeProgram         string `json:"before_program"`
	AfterProgram          string `json:"after_program"`
	TriggerReason         string `json:"trigger_reason"`
	ExecutionAllowed      bool   `json:"execution_allowed"`
	RepositoryWrites      int    `json:"repository_writes"`
	NextOperation         string `json:"next_operation"`
	BlockedBy             []string `json:"blocked_by"`
}

type SourceRevisionEvaluation struct {
	Schema                string      `json:"schema"`
	State                 ReplayState `json:"state"`
	Reason                string      `json:"reason"`
	NextOperation         string      `json:"next_operation"`
	BlockedBy             []string    `json:"blocked_by"`
	SourceDigest          string      `json:"source_digest"`
	CandidateSourceDigest string      `json:"candidate_source_digest"`
	Activity              string      `json:"activity"`
	InputDigest           string      `json:"input_digest"`
	BaselineExecution     Execution   `json:"baseline_execution"`
	CandidateExecution    Execution   `json:"candidate_execution"`
	BaselineFailure       *Failure    `json:"baseline_failure,omitempty"`
	CandidateExecuted     bool        `json:"candidate_executed"`
	Accepted              bool        `json:"accepted"`
	RepositoryWrites      int         `json:"repository_writes"`
}

// ProposeSourceRevision creates an exact, external candidate. It never edits
// the supplied source and never grants execution or adoption authority.
func ProposeSourceRevision(filename string, source []byte, request SourceRevisionRequest) ([]byte, SourceRevision, error) {
	if !validDigest(request.SourceDigest) || request.SourceDigest != digestBytes(source) {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "match-source-digest", "source digest does not match the supplied source")
	}
	if strings.TrimSpace(request.Activity) == "" || request.ExpectedProgram == "" || request.ReplacementProgram == "" || request.TriggerReason == "" {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "require-explicit-revision", "activity, expected program, replacement program and trigger reason are required")
	}
	if request.ExpectedProgram == request.ReplacementProgram {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "reject-noop-revision", request.Activity)
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() || file == nil {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "parse-source", diagnostics.Error().Error())
	}
	var target *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != request.Activity {
			continue
		}
		if target != nil {
			return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "reject-ambiguous-activity", request.Activity)
		}
		target = activity
	}
	if target == nil || !target.ValueProgramPresent || target.ValueProgram != request.ExpectedProgram {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "match-expected-program", request.Activity)
	}
	start, end := target.ValueProgramSpan.Start.Offset, target.ValueProgramSpan.End.Offset
	if start < 0 || end <= start || end > len(source) {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "validate-program-span", request.Activity)
	}
	quoted := string(source[start:end])
	decoded, err := strconv.Unquote(quoted)
	if err != nil || decoded != request.ExpectedProgram {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "match-source-token", request.Activity)
	}
	replacement := strconv.Quote(request.ReplacementProgram)
	candidate := make([]byte, 0, len(source)+len(replacement)-len(quoted))
	candidate = append(candidate, source[:start]...)
	candidate = append(candidate, replacement...)
	candidate = append(candidate, source[end:]...)
	candidateFile, candidateDiagnostics := syntax.ParseFile(filename, string(candidate))
	if candidateDiagnostics.HasErrors() || candidateFile == nil {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "parse-candidate-source", candidateDiagnostics.Error().Error())
	}
	revision := SourceRevision{
		Schema: SourceRevisionSchema, SourceDigest: request.SourceDigest,
		CandidateSourceDigest: digestBytes(candidate), Activity: request.Activity,
		BeforeProgram: request.ExpectedProgram, AfterProgram: request.ReplacementProgram,
		TriggerReason: request.TriggerReason, ExecutionAllowed: false, RepositoryWrites: 0,
		NextOperation: "EVALUATE_SOURCE_REVISION_INDEPENDENTLY", BlockedBy: []string{"independent_evaluation"},
	}
	revision.CandidateID = "gooo://source-revision/" + digestValue(revision)[len("sha256:") : len("sha256:")+16]
	return candidate, revision, nil
}

// EvaluateSourceRevision compares baseline and candidate execution in the
// same process boundary. It returns a state receipt for every outcome and
// never writes source or repository state.
func EvaluateSourceRevision(revision SourceRevision, baselineFilename string, baselineSource []byte, candidateFilename string, candidateSource []byte, activity string, input int64) SourceRevisionEvaluation {
	evaluation := SourceRevisionEvaluation{
		Schema: SourceRevisionEvaluationSchema, State: ReplayUnknown,
		Reason: "SOURCE_REVISION_EVALUATION_UNKNOWN", NextOperation: "REPAIR_SOURCE_REVISION_EVALUATION_INPUT",
		BlockedBy: []string{"revision_receipt", "source_digests", "activity"},
		SourceDigest: digestBytes(baselineSource), CandidateSourceDigest: digestBytes(candidateSource),
		Activity: activity, InputDigest: digestValue(map[string]int64{activity: input}), RepositoryWrites: 0,
	}
	if revision.Schema != SourceRevisionSchema || revision.ExecutionAllowed || revision.RepositoryWrites != 0 || revision.SourceDigest != evaluation.SourceDigest || revision.CandidateSourceDigest != evaluation.CandidateSourceDigest || revision.Activity != activity || revision.TriggerReason == "" {
		return evaluation
	}
	baselinePlan, err := CompilePlan(baselineFilename, baselineSource)
	if err != nil {
		return refuteSourceRevision(evaluation, "SOURCE_REVISION_BASELINE_NOT_EXECUTABLE", "PRESERVE_BASELINE_COMPILE_FAILURE")
	}
	evaluation.BaselineExecution, err = baselinePlan.Execute(map[string]int64{activity: input})
	if err == nil {
		return refuteSourceRevision(evaluation, "SOURCE_REVISION_BASELINE_NOT_FAILED", "PRESERVE_NON_FAILING_BASELINE")
	}
	failure, ok := FailureOf(err)
	if !ok || failure.Code != revision.TriggerReason {
		return refuteSourceRevision(evaluation, "SOURCE_REVISION_BASELINE_FAILURE_MISMATCH", "PRESERVE_BASELINE_COUNTEREXAMPLE")
	}
	evaluation.BaselineFailure = &failure
	candidatePlan, err := CompilePlan(candidateFilename, candidateSource)
	if err != nil {
		return refuteSourceRevision(evaluation, "SOURCE_REVISION_CANDIDATE_NOT_RECOVERED", "PRESERVE_CANDIDATE_COMPILE_FAILURE")
	}
	evaluation.CandidateExecution, err = candidatePlan.Execute(map[string]int64{activity: input})
	evaluation.CandidateExecuted = true
	if err != nil {
		return refuteSourceRevision(evaluation, "SOURCE_REVISION_CANDIDATE_NOT_RECOVERED", "PRESERVE_CANDIDATE_COUNTEREXAMPLE")
	}
	evaluation.State = ReplayClosed
	evaluation.Reason = "SOURCE_REVISION_RECOVERED_BASELINE_FAILURE"
	evaluation.NextOperation = "RUN_ACCEPTED_SOURCE_REVISION"
	evaluation.BlockedBy = nil
	evaluation.Accepted = true
	return evaluation
}

func refuteSourceRevision(evaluation SourceRevisionEvaluation, reason, next string) SourceRevisionEvaluation {
	evaluation.State = ReplayRefuted
	evaluation.Reason = reason
	evaluation.NextOperation = next
	evaluation.BlockedBy = nil
	return evaluation
}

func (evaluation SourceRevisionEvaluation) String() string {
	return fmt.Sprintf("%s: %s", evaluation.State, evaluation.Reason)
}
