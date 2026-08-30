package selfimprovementloop

const (
	Schema             = "gooo/self-improvement-minimal-loop/v1"
	GraphSchemaVersion = "gooo-graph/v1"
	DecisionClosed     = "CLOSED"
	DecisionUnknown    = "UNKNOWN"
	DecisionRefuted    = "REFUTED"
)

var fixedCells = [...]string{
	"OBSERVE_BASELINE",
	"DECLARE_TARGET",
	"PIN_SCOPE",
	"BIND_META_ACTIVITY",
	"PROPOSE_TRANSFORMATION",
	"PREDICT_EFFECT",
	"BUILD_COUNTEREXAMPLE",
	"EXECUTE_CI",
	"CAPTURE_RECEIPT",
	"COMPARE_EXACT_PAIR",
	"HUMAN_DECISION",
	"PROPAGATE_OR_REFUTE",
}

// SemanticCells returns the immutable order of the loop's meaning cells.
func SemanticCells() []string {
	return append([]string(nil), fixedCells[:]...)
}

type Graph struct {
	SchemaVersion string      `json:"schema_version"`
	GraphHash     string      `json:"graph_hash"`
	SourceDigest  string      `json:"source_digest"`
	Nodes         []GraphNode `json:"nodes"`
}

type GraphNode struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type Input struct {
	Schema          string                 `json:"schema"`
	Scenario        string                 `json:"scenario"`
	SourceDigest    string                 `json:"source_digest"`
	ToolchainDigest string                 `json:"toolchain_digest"`
	Baseline        BaselineObservation    `json:"baseline"`
	Target          TargetDeclaration      `json:"target"`
	Scope           ScopePin               `json:"scope"`
	Transformation  TransformationProposal `json:"transformation"`
	Prediction      EffectPrediction       `json:"prediction"`
	Counterexample  CounterexampleResult   `json:"counterexample"`
	CI              CIResult               `json:"ci"`
	Receipt         ReceiptInput           `json:"receipt"`
	Pair            ExactPair              `json:"pair"`
	Human           HumanDecision          `json:"human"`
}

type BaselineObservation struct {
	Present bool   `json:"present"`
	Metric  string `json:"metric"`
	Value   int64  `json:"value"`
}

type TargetDeclaration struct {
	Present bool   `json:"present"`
	Metric  string `json:"metric"`
	Value   int64  `json:"value"`
}

type ScopePin struct {
	Paths []string `json:"paths"`
}

type TransformationProposal struct {
	Present            bool   `json:"present"`
	Patch              string `json:"patch"`
	OutputMode         string `json:"output_mode"`
	RepositoryMutation bool   `json:"repository_mutation"`
}

type EffectPrediction struct {
	Present bool   `json:"present"`
	Metric  string `json:"metric"`
	Before  int64  `json:"before"`
	After   int64  `json:"after"`
}

type CounterexampleResult struct {
	Checked  bool   `json:"checked"`
	Found    bool   `json:"found"`
	Evidence string `json:"evidence"`
}

type CIResult struct {
	Executed bool   `json:"executed"`
	Passed   bool   `json:"passed"`
	Evidence string `json:"evidence"`
}

type ReceiptInput struct {
	Captured bool   `json:"captured"`
	Digest   string `json:"digest"`
}

type ExactPair struct {
	Before []MetricSample `json:"before"`
	After  []MetricSample `json:"after"`
}

type MetricSample struct {
	Scenario        string `json:"scenario"`
	SourceDigest    string `json:"source_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	Metric          string `json:"metric"`
	Value           int64  `json:"value"`
}

type HumanDecision struct {
	Decision string `json:"decision"`
}

// UnknownState is deliberately not compressed into a message or score.
type UnknownState struct {
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	UnknownClass  string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
	BlockedBy     string `json:"blocked_by"`
}

type ActivityBinding struct {
	Cell       string `json:"cell"`
	Activity   string `json:"activity"`
	ActivityID string `json:"activity_id"`
}

type CellResult struct {
	Cell       string        `json:"cell"`
	Activity   string        `json:"activity"`
	ActivityID string        `json:"activity_id"`
	Decision   string        `json:"decision"`
	Reason     string        `json:"reason"`
	Unknown    *UnknownState `json:"unknown,omitempty"`
}

type Report struct {
	Schema       string            `json:"schema"`
	Scenario     string            `json:"scenario"`
	SourceDigest string            `json:"source_digest"`
	ToolchainDigest string         `json:"toolchain_digest"`
	Decision     string            `json:"decision"`
	Reason       string            `json:"reason"`
	GraphHash    string            `json:"graph_hash"`
	Cells        []CellResult      `json:"cells"`
	Bindings     []ActivityBinding `json:"bindings"`
	Unknowns     []UnknownState    `json:"unknowns,omitempty"`
	PairMatched  bool              `json:"pair_matched"`
	ReportDigest string            `json:"report_digest"`
}

type PatchProposal struct {
	Schema             string `json:"schema"`
	Scenario           string `json:"scenario"`
	OutputMode         string `json:"output_mode"`
	RepositoryMutation bool   `json:"repository_mutation"`
	Patch              string `json:"patch"`
}

type EvidenceRecord struct {
	Cell     string        `json:"cell"`
	Decision string        `json:"decision"`
	Reason   string        `json:"reason"`
	Unknown  *UnknownState `json:"unknown,omitempty"`
}

type EvidenceBundle struct {
	Schema         string           `json:"schema"`
	Scenario       string           `json:"scenario"`
	SourceDigest   string           `json:"source_digest"`
	ToolchainDigest string          `json:"toolchain_digest"`
	Decision       string           `json:"decision"`
	Cells          []EvidenceRecord `json:"cells"`
	Unknowns       []UnknownState   `json:"unknowns,omitempty"`
	Pair           ExactPair        `json:"pair"`
	GraphHash      string           `json:"graph_hash"`
	EvidenceDigest string           `json:"evidence_digest"`
}

type Dossier struct {
	Schema         string         `json:"schema"`
	Scenario       string         `json:"scenario"`
	SourceDigest   string         `json:"source_digest"`
	ToolchainDigest string        `json:"toolchain_digest"`
	Decision       string         `json:"decision"`
	GraphHash      string         `json:"graph_hash"`
	ReportDigest   string         `json:"report_digest"`
	PatchProposal  PatchProposal  `json:"patch_proposal"`
	EvidenceDigest string         `json:"evidence_digest"`
	Unknowns       []UnknownState `json:"unknowns,omitempty"`
	DossierDigest  string         `json:"dossier_digest"`
}

type Artifacts struct {
	Report        Report
	PatchProposal PatchProposal
	Evidence      EvidenceBundle
	Dossier       Dossier
}
