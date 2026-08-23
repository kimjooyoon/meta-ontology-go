package operationconformance

const (
	ReportSchema            = "gooo/operation-conformance-report/v1"
	ContractSchema          = "gooo/operation-conformance-contract/v1"
	ContractID              = "gooo/source-splitter/split-go-conformance/v1"
	OperationID             = "gooo/meta/generation/SplitGo"
	DenominatorVersion      = "split-go-indicators/v1"
	BehavioralCorpusSchema  = "gooo/split-go-behavioral-corpus/v1"
	BehavioralCorpusVersion = "split-go-behavior-cases/v1"
)

type Decision string

const (
	DecisionPass    Decision = "PASS"
	DecisionFail    Decision = "FAIL"
	DecisionUnknown Decision = "UNKNOWN"
	DecisionBlock   Decision = "BLOCK"
)

type BuildContext struct {
	GOOS       string   `json:"goos"`
	GOARCH     string   `json:"goarch"`
	CgoEnabled bool     `json:"cgo_enabled"`
	BuildTags  []string `json:"build_tags,omitempty"`
}

type FileEvidence struct {
	Path string `json:"path"`
	Data []byte `json:"data"`
}

type WriteEvent struct {
	Sequence  int    `json:"sequence"`
	Kind      string `json:"kind"`
	Target    string `json:"target"`
	Temporary string `json:"temporary,omitempty"`
	Success   bool   `json:"success"`
}

type WriteReceipt struct {
	Complete                     bool         `json:"complete"`
	ExecutionSucceeded           bool         `json:"execution_succeeded"`
	DeclaredTargets              []string     `json:"declared_targets"`
	Events                       []WriteEvent `json:"events"`
	WritesOutsideDeclaredTargets int          `json:"writes_outside_declared_targets"`
	TemporaryFilesRemaining      int          `json:"temporary_files_remaining"`
}

type SplitGoEvidence struct {
	ExpectedHeadSHA  string         `json:"expected_head_sha"`
	OperationID      string         `json:"operation_id"`
	EvidenceComplete bool           `json:"evidence_complete"`
	Source           FileEvidence   `json:"source"`
	Candidates       []FileEvidence `json:"candidates"`
	BuildContexts    []BuildContext `json:"build_contexts"`
	Write            WriteReceipt   `json:"write_receipt"`
}

type IndicatorDefinition struct {
	ID     string `json:"id"`
	Role   string `json:"role"`
	Route  string `json:"route"`
	RuleID string `json:"rule_id"`
}
