package transformationeffect

import (
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
	effects    []Effect
	receipts   generation.ReceiptReport
	provenance generation.ArtifactProvenance
	baseline   workspace.State
	final      workspace.State
	patch      workspace.Patch
}

func effectFor(action generation.Action, before, after workspace.State, changes []workspace.Change, evidence, receipt string) Effect {
	return Effect{ActionIndicatorID: action.IndicatorID, MetricID: string(action.MetricID), Subject: action.Subject,
		SubjectKind: string(action.SubjectKind), Operation: string(action.Operation), Executor: action.Executor,
		Evaluator: action.Evaluator, ProofChoice: string(action.ProofChoice), BeforeTreeDigest: before.Digest,
		AfterTreeDigest: after.Digest, ChangedPathCount: len(changes), ChangedPathDigest: hashJSON(changes),
		ResidualActionable: 0, EvaluatorEvidence: evidence, ReceiptDigest: receipt, Status: "APPLIED"}
}
