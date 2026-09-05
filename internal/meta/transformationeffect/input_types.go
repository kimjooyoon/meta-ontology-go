package transformationeffect

import (
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect/workspace"
)

type Options struct {
	Root           string
	MetricsPath    string
	PlanPath       string
	ExecutionPath  string
	ReceiptsPath   string
	ProvenancePath string
	ExpectedSHA    string
	ProgressWriter io.Writer
	InvocationID   string
}

// operationProgressEvent is a diagnostic-only control-flow trace. Boundary
// values record command entry/return and never assert semantic success.
type operationProgressEvent struct {
	Schema                       string `json:"schema"`
	HeadSHA                      string `json:"head_sha"`
	InvocationID                 string `json:"invocation_id"`
	Sequence                     int    `json:"sequence"`
	ActionIndicatorID            string `json:"action_indicator_id"`
	Operation                    string `json:"operation"`
	Activity                     string `json:"activity"`
	Executor                     string `json:"executor"`
	Subject                      string `json:"subject"`
	SubjectKind                  string `json:"subject_kind"`
	InputContractSourceDigest    string `json:"input_contract_source_digest"`
	InputContractSemanticDigest  string `json:"input_contract_semantic_digest"`
	Phase                        string `json:"phase"`
	Boundary                     string `json:"boundary"`
	ReturnError                  string `json:"return_error,omitempty"`
}

type ArtifactDigests struct {
	SourceMetrics string `json:"source_metrics"`
	Plan          string `json:"plan"`
	Execution     string `json:"execution"`
	Receipts      string `json:"receipts"`
	Provenance    string `json:"provenance"`
}

type inputSet struct {
	metrics    linecaps.LineMetricsReport
	plan       generation.Plan
	execution  generation.ExecutionManifest
	receipts   generation.ReceiptReport
	provenance generation.ArtifactProvenance
	digests    ArtifactDigests
}

type Result struct {
	Ledger     Ledger
	Patch      workspace.Patch
	Receipts   generation.ReceiptReport
	Provenance generation.ArtifactProvenance
}

type executionResult struct {
	effects                   []Effect
	failures                  []generation.ObservationFailure
	receipts                  generation.ReceiptReport
	provenance                generation.ArtifactProvenance
	selectedPlanOperations    int
	boundExecutorOperations   int
	unboundExecutorOperations int
	baseline                  workspace.State
	final                     workspace.State
	patch                     workspace.Patch
}

func effectFor(action generation.Action, before, after workspace.State, changes []workspace.Change, evidence, receipt string) Effect {
	return Effect{ActionIndicatorID: action.IndicatorID, MetricID: string(action.MetricID), Subject: action.Subject,
		SubjectKind: string(action.SubjectKind), Operation: string(action.Operation), Activity: action.Activity,
		Output: action.Output, Executor: action.Executor,
		Evaluator: action.Evaluator, ProofChoice: string(action.ProofChoice), BeforeTreeDigest: before.Digest,
		AfterTreeDigest: after.Digest, ChangedPathCount: len(changes), ChangedPathDigest: hashJSON(changes),
		ResidualActionable: 0, EvaluatorEvidence: evidence, ReceiptDigest: receipt, Status: "APPLIED"}
}

func effectForFailure(action generation.Action, before workspace.State, failure generation.ObservationFailure) Effect {
	changes := []workspace.Change{}
	evidence := hashJSON(failure)
	return Effect{ActionIndicatorID: action.IndicatorID, MetricID: string(action.MetricID), Subject: action.Subject,
		SubjectKind: string(action.SubjectKind), Operation: string(action.Operation), Activity: action.Activity,
		Output: action.Output, Executor: action.Executor,
		Evaluator: action.Evaluator, ProofChoice: string(action.ProofChoice), BeforeTreeDigest: before.Digest,
		AfterTreeDigest: before.Digest, ChangedPathDigest: hashJSON(changes), EvaluatorEvidence: evidence,
		ReceiptDigest: evidence, Status: failure.Decision}
}
