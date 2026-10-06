package main

const (
	ReceiptSchema = "gooo://meta/domain-completeness/receipt/v1"
	ProfileID     = "gooo://meta/domain-completeness/profile/gooo-language-utility-v1"
)

type Entity struct {
	Name string
	ID   string
}

type Activity struct {
	Name         string
	Inputs       []string
	Output       string
	ValueProgram string
}

type EvidenceRef struct {
	Role   string `json:"role"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type Frontier struct {
	Unit          string `json:"unit"`
	Stage         string `json:"stage"`
	Reason        string `json:"reason"`
	NextOperation string `json:"next_operation"`
}

type Dimension struct {
	ID              string        `json:"id"`
	MetricID        string        `json:"metric_id"`
	Status          string        `json:"status"`
	Numerator       int           `json:"numerator"`
	Denominator     int           `json:"denominator"`
	Unit            string        `json:"unit"`
	UnknownUnits    int           `json:"unknown_units"`
	RefutedUnits    int           `json:"refuted_units"`
	Evidence        []EvidenceRef `json:"evidence"`
	FirstUnresolved *Frontier     `json:"first_unresolved,omitempty"`
}

type Snapshot struct {
	Repository  string `json:"repository"`
	SubjectSHA  string `json:"subject_sha"`
	WorkflowRun int64  `json:"workflow_run_id"`
	RunAttempt  int    `json:"run_attempt"`
	Toolchain   string `json:"toolchain"`
	Evaluator   string `json:"evaluator"`
}

type SystemBudget struct {
	MaximumEvidenceFiles    int   `json:"maximum_evidence_files"`
	MaximumEvidenceBytes    int64 `json:"maximum_evidence_bytes"`
	MaximumRepositoryWrites int   `json:"maximum_repository_writes"`
	MaximumHumanActions     int   `json:"maximum_human_actions"`
}

type SystemCost struct {
	EvidenceFiles    int   `json:"evidence_files"`
	EvidenceBytes    int64 `json:"evidence_bytes"`
	RepositoryWrites int   `json:"repository_writes"`
	HumanActions     int   `json:"human_actions"`
}

type Report struct {
	Schema          string       `json:"schema"`
	ProfileID       string       `json:"profile_id"`
	SubjectSHA      string       `json:"subject_sha"`
	Decision        string       `json:"decision"`
	Reason          string       `json:"reason"`
	NextOperation   string       `json:"next_operation"`
	FirstUnresolved *Frontier    `json:"first_unresolved,omitempty"`
	Contract        ContractRef  `json:"contract"`
	Generated       GeneratedRef `json:"generated"`
	Snapshot        Snapshot     `json:"snapshot"`
	RequiredScope   []string     `json:"required_scope"`
	ExcludedScope   []string     `json:"excluded_scope"`
	Dimensions      []Dimension  `json:"dimensions"`
	Summary         Summary      `json:"summary"`
	Investment      Investment   `json:"investment"`
	Digest          string       `json:"digest"`
}

type ContractRef struct {
	Path          string `json:"path"`
	Digest        string `json:"digest"`
	SemanticHash  string `json:"semantic_hash"`
	ReceiptSchema string `json:"receipt_schema"`
}

type GeneratedRef struct {
	Path           string `json:"path"`
	Digest         string `json:"digest"`
	SemanticHash   string `json:"semantic_hash"`
	SemanticsEqual bool   `json:"semantics_equal"`
	ReceiptSchema  string `json:"receipt_schema"`
}

type Summary struct {
	DimensionsTotal int `json:"dimensions_total"`
	Pass            int `json:"pass"`
	Progress        int `json:"progress"`
	Unknown         int `json:"unknown"`
	FailClosed      int `json:"fail_closed"`
}

type Investment struct {
	Budget           SystemBudget `json:"system_budget"`
	Observed         SystemCost   `json:"observed_system_cost"`
	ComparisonStatus string       `json:"comparison_status"`
	Comparison       *Comparison  `json:"comparison,omitempty"`
}

type Comparison struct {
	BaselineSubject string           `json:"baseline_subject_sha"`
	BaselineDigest  string           `json:"baseline_receipt_digest"`
	Status          string           `json:"status"`
	Dimensions      []DimensionDelta `json:"dimensions"`
}

type DimensionDelta struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	NumeratorDelta int    `json:"numerator_delta"`
	Denominator    int    `json:"denominator"`
	BaselineStatus string `json:"baseline_status"`
	CurrentStatus  string `json:"current_status"`
}

type ProfileModel struct {
	Package      string
	Namespace    string
	Entities     map[string]Entity
	Activities   map[string]Activity
	SemanticHash string
	SourceDigest string
}
