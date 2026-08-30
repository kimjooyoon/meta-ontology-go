package governancesnapshot

const (
	ReportSchema   = "gooo/live-governance-snapshot/v1"
	ContractSchema = "gooo/live-governance-snapshot-contract/v1"
	SnapshotSchema = "gooo/live-governance-snapshot-input/v1"
	GraphSchema    = "gooo-graph/v1"
	DecisionClosed = "CLOSED"
	DecisionPass   = "PASS"
	DecisionRefuted = "REFUTED"
	DecisionUnknown = "UNKNOWN"
	ResolutionExact = "EXACT"
	ResolutionLower = "LOWER"
)

type Contract struct {
	Schema          string             `json:"schema"`
	ID              string             `json:"id"`
	Source          SourceContract     `json:"source"`
	Expected        ExpectedContract   `json:"expected"`
	Cells           []CellSpec         `json:"cells"`
	GraphProgram    string             `json:"graph_program"`
	NotClaimed      []string           `json:"not_claimed"`
}

type SourceContract struct {
	Documentation []string       `json:"documentation"`
	APIVersions  map[string]string `json:"api_versions"`
	Endpoints    []EndpointSpec `json:"endpoints"`
}

type EndpointSpec struct {
	ID       string `json:"id"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	API      string `json:"api_version"`
	Payload  string `json:"payload"`
}

type ExpectedContract struct {
	DefaultBranch     string              `json:"default_branch"`
	Repository        string              `json:"repository"`
	StatusChecks      []BranchExpectation `json:"status_checks"`
	Rulesets          []RulesetExpectation `json:"rulesets"`
	RequiredRulesetState string           `json:"required_ruleset_state"`
}

type BranchExpectation struct {
	Branch     string   `json:"branch"`
	Protected  bool     `json:"protected"`
	Enforcement string  `json:"enforcement"`
	Contexts   []string `json:"contexts"`
}

type RulesetExpectation struct {
	Branch      string `json:"branch"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Enforcement string `json:"enforcement"`
}

type CellSpec struct {
	ID            string `json:"id"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Indicator     string `json:"indicator"`
	Activity      string `json:"activity"`
	InputID       string `json:"input_id"`
	OutputID      string `json:"output_id"`
}

type Snapshot struct {
	Schema               string               `json:"schema"`
	Repository           string               `json:"repository"`
	HeadSHA              string               `json:"head_sha"`
	Requests             []RequestObservation `json:"requests"`
	RepositoryWrites     int                `json:"repository_writes"`
	BranchSettingWrites  int                `json:"branch_setting_writes"`
	LocalTestExecutions  int                `json:"local_test_executions"`
	CrossProjectGates    int                `json:"cross_project_required_gates"`
}

type RequestObservation struct {
	ID            string `json:"id"`
	Method        string `json:"method"`
	URL           string `json:"url"`
	APIVersion    string `json:"api_version"`
	PayloadPath   string `json:"payload_path"`
	State         string `json:"state"`
	PayloadDigest string `json:"payload_digest"`
}

type LoadedSnapshot struct {
	Snapshot
	Payloads map[string][]byte `json:"-"`
}

type Report struct {
	Schema               string              `json:"schema"`
	ContractID           string              `json:"contract_id"`
	Repository           string              `json:"repository"`
	HeadSHA              string              `json:"head_sha"`
	DefaultBranch        string              `json:"default_branch"`
	Decision             string              `json:"decision"`
	Resolution           string              `json:"resolution"`
	Reason               string              `json:"reason"`
	SettingsHealthy      bool                `json:"settings_healthy"`
	PromotionAuthorized  bool                `json:"promotion_authorized"`
	Source               SourceEvidence      `json:"source"`
	Branches             []BranchEvidence    `json:"branches"`
	Rulesets             []RulesetEvidence   `json:"rulesets"`
	Cells                []CellObservation   `json:"cells"`
	Summary              Summary             `json:"summary"`
	Cases                []CanonicalCase     `json:"canonical_cases"`
	Unknowns             []Unknown           `json:"unknowns"`
	Graph                GraphEvidence       `json:"graph"`
	Replay               ReplayEvidence      `json:"replay"`
	Authority            AuthorityCounters   `json:"authority"`
	RepositoryWrites     int                 `json:"repository_writes"`
	BranchSettingWrites  int                 `json:"branch_setting_writes"`
	LocalTestExecutions  int                 `json:"local_test_executions"`
	CrossProjectGates    int                 `json:"cross_project_required_gates"`
	Improvement          string              `json:"improvement"`
	ReportDigest         string              `json:"report_digest"`
	HumanReport          string              `json:"human_report"`
}

type Summary struct {
	CellsTotal       int `json:"cells_total"`
	ClosedCells      int `json:"closed_cells"`
	UnknownCells     int `json:"unknown_cells"`
	RefutedCells     int `json:"refuted_cells"`
	FoundationCells  int `json:"foundation_cells"`
	CoherenceCells   int `json:"coherence_cells"`
	RegressionCells  int `json:"regression_cells"`
}

type SourceEvidence struct {
	Documentation []string              `json:"documentation"`
	APIVersions  map[string]string      `json:"api_versions"`
	Requests     []RequestObservation   `json:"requests"`
	Payloads     []PayloadEvidence      `json:"payloads"`
}

type PayloadEvidence struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Digest string `json:"normalized_digest,omitempty"`
}

type BranchEvidence struct {
	Branch       string   `json:"branch"`
	CommitSHA    string   `json:"commit_sha"`
	Protected    bool     `json:"protected"`
	Enforcement  string   `json:"status_enforcement"`
	Contexts     []string `json:"required_contexts"`
}

type RulesetEvidence struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Target      string `json:"target"`
	Enforcement string `json:"enforcement"`
}

type CellObservation struct {
	ID            string   `json:"id"`
	MetaOperation string   `json:"meta_operation"`
	ProofChoice   string   `json:"proof_choice"`
	Indicator     string   `json:"indicator"`
	Activity      string   `json:"activity"`
	InputID       string   `json:"input_id"`
	OutputID      string   `json:"output_id"`
	Observed      string   `json:"observed"`
	Expected      string   `json:"expected"`
	Decision      string   `json:"decision"`
	Resolution    string   `json:"resolution"`
	Reason        string   `json:"reason,omitempty"`
	Counterexample string  `json:"counterexample,omitempty"`
	Unknown       *Unknown `json:"unknown,omitempty"`
	EvidenceDigest string  `json:"evidence_digest"`
}

type Unknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type CanonicalCase struct {
	ID            string   `json:"id"`
	Decision      string   `json:"decision"`
	Resolution    string   `json:"resolution"`
	Reason        string   `json:"reason"`
	Expected      string   `json:"expected,omitempty"`
	Observed      string   `json:"observed,omitempty"`
	Counterexample string  `json:"counterexample,omitempty"`
	Unknown       *Unknown `json:"unknown,omitempty"`
}

type GraphEvidence struct {
	Schema        string        `json:"schema"`
	ProgramDigest string        `json:"program_digest"`
	GraphHash     string        `json:"graph_hash"`
	NodeCount     int           `json:"node_count"`
	RelationCount int           `json:"relation_count"`
	ActivityCount int           `json:"activity_count"`
	BindingCount  int           `json:"binding_count"`
	Activities    []string      `json:"activities"`
	Bindings      []GraphBinding `json:"bindings"`
}

type GraphBinding struct {
	CellID          string `json:"cell_id"`
	ActivityID      string `json:"activity_id"`
	InputID         string `json:"input_id"`
	OutputID        string `json:"output_id"`
	UsedEdgeCount   int    `json:"used_edge_count"`
	GeneratedCount  int    `json:"generated_edge_count"`
}

type RawGraph struct {
	SchemaVersion string         `json:"schema_version"`
	SourceDigest  string         `json:"source_digest"`
	GraphHash     string         `json:"graph_hash"`
	Nodes         []GraphNode    `json:"nodes"`
	Relations     []GraphRelation `json:"relations"`
}

type GraphNode struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type GraphRelation struct {
	Subject  string `json:"subject"`
	Predicate string `json:"predicate"`
	Object   string `json:"object"`
}

type ReplayEvidence struct {
	InputDigest     string `json:"input_digest"`
	ProjectionDigest string `json:"projection_digest"`
	ReplayEqual     bool   `json:"replay_equal"`
}

type AuthorityCounters struct {
	RepositoryWrites    int `json:"repository_writes"`
	BranchSettingWrites int `json:"branch_setting_writes"`
	LocalTestExecutions int `json:"local_test_executions"`
	CrossProjectGates   int `json:"cross_project_required_gates"`
}
