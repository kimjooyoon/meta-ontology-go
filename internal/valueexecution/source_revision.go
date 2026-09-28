package valueexecution

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const SourceRevisionSchema = "gooo/value-execution-source-revision/v1"
const SourceRevisionEvaluationSchema = "gooo/value-execution-source-revision-evaluation/v1"
const SourceRevisionContractScope = "EXPLICIT_INPUT_SET"

type SourceRevisionRequest struct {
	SourceDigest       string              `json:"source_digest"`
	Activity           string              `json:"activity"`
	ExpectedProgram    string              `json:"expected_program"`
	ReplacementProgram string              `json:"replacement_program"`
	TriggerReason      string              `json:"trigger_reason"`
	AnalysisProvenance *AnalysisProvenance `json:"analysis_provenance,omitempty"`
}

type AnalysisProvenance struct {
	SourceDigest    string `json:"source_digest"`
	ProfileDigest   string `json:"profile_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	ContractDigest  string `json:"contract_digest"`
}

type SourceRevision struct {
	Schema                string              `json:"schema"`
	CandidateID           string              `json:"candidate_id"`
	SourceDigest          string              `json:"source_digest"`
	CandidateSourceDigest string              `json:"candidate_source_digest"`
	Activity              string              `json:"activity"`
	BeforeProgram         string              `json:"before_program"`
	AfterProgram          string              `json:"after_program"`
	TriggerReason         string              `json:"trigger_reason"`
	ExecutionAllowed      bool                `json:"execution_allowed"`
	RepositoryWrites      int                 `json:"repository_writes"`
	NextOperation         string              `json:"next_operation"`
	BlockedBy             []string            `json:"blocked_by"`
	RepairHandoffDigest   string              `json:"repair_handoff_digest,omitempty"`
	RepairCandidateID     string              `json:"repair_candidate_id,omitempty"`
	AnalysisProvenance    *AnalysisProvenance `json:"analysis_provenance,omitempty"`
}

type SourceRevisionContract struct {
	Scope           string          `json:"scope"`
	Inputs          []int64         `json:"inputs"`
	ExpectedOutputs map[int64]int64 `json:"expected_outputs,omitempty"`
}

type SourceRevisionEvaluation struct {
	Schema                  string              `json:"schema"`
	State                   ReplayState         `json:"state"`
	Reason                  string              `json:"reason"`
	NextOperation           string              `json:"next_operation"`
	BlockedBy               []string            `json:"blocked_by"`
	SourceDigest            string              `json:"source_digest"`
	CandidateSourceDigest   string              `json:"candidate_source_digest"`
	Activity                string              `json:"activity"`
	InputDigest             string              `json:"input_digest"`
	BaselineExecution       Execution           `json:"baseline_execution"`
	CandidateExecution      Execution           `json:"candidate_execution"`
	BaselineFailure         *Failure            `json:"baseline_failure,omitempty"`
	CandidateExecuted       bool                `json:"candidate_executed"`
	Scope                   string              `json:"scope"`
	CounterexampleRecovered bool                `json:"counterexample_recovered"`
	ContractInputs          []int64             `json:"contract_inputs,omitempty"`
	ContractExpectedOutputs map[int64]int64     `json:"contract_expected_outputs,omitempty"`
	ContractDigest          string              `json:"contract_digest,omitempty"`
	AnalysisProvenance      *AnalysisProvenance `json:"analysis_provenance,omitempty"`
	ContractPreservation    bool                `json:"contract_preservation"`
	RegressionEvidence      []string            `json:"regression_evidence,omitempty"`
	AdoptionAuthorized      bool                `json:"adoption_authorized"`
	Accepted                bool                `json:"accepted"`
	RepositoryWrites        int                 `json:"repository_writes"`
}

// ProposeSourceRevision creates an exact, external candidate. It never edits
// the supplied source and never grants execution or adoption authority.
func ProposeSourceRevision(filename string, source []byte, request SourceRevisionRequest) ([]byte, SourceRevision, error) {
	if !validDigest(request.SourceDigest) || request.SourceDigest != digestBytes(source) {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "match-source-digest", "source digest does not match the supplied source")
	}
	if request.AnalysisProvenance != nil && !request.AnalysisProvenance.validFor(request.SourceDigest) {
		return nil, SourceRevision{}, failAt(ReasonSourceRevisionInvalid, "PROPOSE", "validate-analysis-provenance", "analysis provenance is not bound to the supplied source")
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
		AnalysisProvenance: cloneAnalysisProvenance(request.AnalysisProvenance),
	}
	revision.CandidateID = "gooo://source-revision/" + digestValue(revision)[len("sha256:"):len("sha256:")+16]
	return candidate, revision, nil
}

// EvaluateSourceRevision compares baseline and candidate execution in the
// same process boundary. It returns a state receipt for every outcome and
// never writes source or repository state.
func EvaluateSourceRevision(revision SourceRevision, baselineFilename string, baselineSource []byte, candidateFilename string, candidateSource []byte, activity string, input int64) SourceRevisionEvaluation {
	evaluation := SourceRevisionEvaluation{
		Schema: SourceRevisionEvaluationSchema, State: ReplayUnknown,
		Reason: "SOURCE_REVISION_EVALUATION_UNKNOWN", NextOperation: "REPAIR_SOURCE_REVISION_EVALUATION_INPUT",
		BlockedBy:    []string{"revision_receipt", "source_digests", "activity"},
		SourceDigest: digestBytes(baselineSource), CandidateSourceDigest: digestBytes(candidateSource),
		Activity: activity, InputDigest: digestValue(map[string]int64{activity: input}), RepositoryWrites: 0,
		AnalysisProvenance: cloneAnalysisProvenance(revision.AnalysisProvenance),
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
	evaluation.Scope = "COUNTEREXAMPLE_RECOVERY"
	evaluation.CounterexampleRecovered = true
	evaluation.Reason = "SOURCE_REVISION_COUNTEREXAMPLE_RECOVERED"
	evaluation.NextOperation = "VERIFY_SOURCE_REVISION_CONTRACT"
	evaluation.BlockedBy = []string{"contract_preservation", "explicit_adoption_decision"}
	evaluation.Accepted = false
	return evaluation
}

func (provenance *AnalysisProvenance) validFor(sourceDigest string) bool {
	return provenance != nil && provenance.SourceDigest == sourceDigest && validDigest(provenance.ProfileDigest) && validDigest(provenance.ToolchainDigest) && validDigest(provenance.ContractDigest)
}

func cloneAnalysisProvenance(provenance *AnalysisProvenance) *AnalysisProvenance {
	if provenance == nil {
		return nil
	}
	copy := *provenance
	return &copy
}

// VerifySourceRevisionContract checks a bounded, explicit set of successful
// baseline inputs. Recovering one counterexample is not evidence that the
// candidate preserves the surrounding contract.
func VerifySourceRevisionContract(revision SourceRevision, evaluation SourceRevisionEvaluation, baselineFilename string, baselineSource []byte, candidateFilename string, candidateSource []byte, activity string, contract SourceRevisionContract) SourceRevisionEvaluation {
	if evaluation.State != ReplayClosed || !evaluation.CounterexampleRecovered || !evaluation.CandidateExecuted || evaluation.RepositoryWrites != 0 || contract.Scope != SourceRevisionContractScope || len(contract.Inputs) == 0 {
		return refuteSourceRevision(evaluation, "SOURCE_REVISION_CONTRACT_UNKNOWN", "REPAIR_SOURCE_REVISION_CONTRACT_INPUT")
	}
	baselinePlan, err := CompilePlan(baselineFilename, baselineSource)
	if err != nil {
		return refuteSourceRevision(evaluation, "SOURCE_REVISION_CONTRACT_BASELINE_NOT_EXECUTABLE", "PRESERVE_BASELINE_CONTRACT")
	}
	candidatePlan, err := CompilePlan(candidateFilename, candidateSource)
	if err != nil {
		return refuteSourceRevision(evaluation, "SOURCE_REVISION_CONTRACT_CANDIDATE_NOT_EXECUTABLE", "PRESERVE_CANDIDATE_CONTRACT")
	}
	evidence := make([]string, 0, len(contract.Inputs))
	for _, input := range contract.Inputs {
		baselineExecution, baselineErr := baselinePlan.Execute(map[string]int64{activity: input})
		if baselineErr != nil {
			return refuteSourceRevision(evaluation, "SOURCE_REVISION_CONTRACT_BASELINE_NOT_PROVEN", "PRESERVE_BASELINE_CONTRACT")
		}
		if expected, ok := contract.ExpectedOutputs[input]; ok {
			baselineResult, resultOK := baselineExecution.Results[activity]
			if !resultOK || baselineResult.Value != expected {
				return refuteSourceRevision(evaluation, "SOURCE_REVISION_CONTRACT_BASELINE_EXPECTATION_MISMATCH", "REVIEW_SOURCE_REVISION_CONTRACT_EXPECTATIONS")
			}
		}
		candidateExecution, candidateErr := candidatePlan.Execute(map[string]int64{activity: input})
		if candidateErr != nil {
			return refuteSourceRevision(evaluation, "SOURCE_REVISION_CONTRACT_CANDIDATE_REGRESSION", "PRESERVE_CANDIDATE_CONTRACT")
		}
		if expected, ok := contract.ExpectedOutputs[input]; ok {
			candidateResult, resultOK := candidateExecution.Results[activity]
			if !resultOK || candidateResult.Value != expected {
				return refuteSourceRevision(evaluation, "SOURCE_REVISION_CONTRACT_CANDIDATE_EXPECTATION_MISMATCH", "PRESERVE_CANDIDATE_CONTRACT")
			}
		}
		if !sameContractResult(baselineExecution, candidateExecution, activity) {
			return refuteSourceRevision(evaluation, "SOURCE_REVISION_CONTRACT_OUTPUT_CHANGED", "PRESERVE_CANDIDATE_CONTRACT")
		}
		evidence = append(evidence, digestValue(map[string]any{
			"activity": activity, "input": input,
			"baseline":  baselineExecution.Results[activity].Value,
			"candidate": candidateExecution.Results[activity].Value,
		}))
	}
	evaluation.Scope = "CONTRACT_PRESERVATION"
	evaluation.ContractInputs = append([]int64(nil), contract.Inputs...)
	if len(contract.ExpectedOutputs) > 0 {
		evaluation.ContractExpectedOutputs = make(map[int64]int64, len(contract.ExpectedOutputs))
		maps.Copy(evaluation.ContractExpectedOutputs, contract.ExpectedOutputs)
	}
	contractDigestInput := map[string]any{"scope": contract.Scope, "inputs": contract.Inputs, "evidence": evidence}
	if len(contract.ExpectedOutputs) > 0 {
		contractDigestInput["expected_outputs"] = contract.ExpectedOutputs
	}
	evaluation.ContractDigest = digestValue(contractDigestInput)
	evaluation.ContractPreservation = true
	evaluation.RegressionEvidence = evidence
	evaluation.AdoptionAuthorized = false
	evaluation.Accepted = true
	evaluation.Reason = "SOURCE_REVISION_CONTRACT_PRESERVED"
	evaluation.NextOperation = "REQUEST_EXPLICIT_SOURCE_REVISION_ADOPTION"
	evaluation.BlockedBy = []string{"explicit_adoption_decision"}
	return evaluation
}

func sameContractResult(baseline, candidate Execution, activity string) bool {
	base, baseOK := baseline.Results[activity]
	candidateResult, candidateOK := candidate.Results[activity]
	return baseOK && candidateOK && base.Scope == candidateResult.Scope && base.ProducerActivity == candidateResult.ProducerActivity && base.OutputEntity == candidateResult.OutputEntity && base.Value == candidateResult.Value
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
