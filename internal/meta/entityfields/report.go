package entityfields

type Unknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type CellObservation struct {
	CellSpec
	Observed       string   `json:"observed"`
	Expected       string   `json:"expected"`
	Decision       string   `json:"decision"`
	Resolution     string   `json:"resolution"`
	Reason         string   `json:"reason,omitempty"`
	EvidenceDigest string   `json:"evidence_digest"`
	Unknown        *Unknown `json:"unknown,omitempty"`
}

type Summary struct {
	CellsTotal      int `json:"cells_total"`
	ClosedCells     int `json:"closed_cells"`
	UnknownCells    int `json:"unknown_cells"`
	RefutedCells    int `json:"refuted_cells"`
	FoundationCells int `json:"foundation_cells"`
	CoherenceCells  int `json:"coherence_cells"`
	RegressionCells int `json:"regression_cells"`
	DriverCells     int `json:"driver_cells"`
	OutcomeCells    int `json:"outcome_cells"`
	GuardrailCells  int `json:"guardrail_cells"`
}

type Counterexample struct {
	ID             string   `json:"id"`
	Decision       string   `json:"decision"`
	Resolution     string   `json:"resolution"`
	Reason         string   `json:"reason"`
	Expected       string   `json:"expected,omitempty"`
	Observed       string   `json:"observed,omitempty"`
	InputDigest    string   `json:"input_digest,omitempty"`
	OutputDigest   string   `json:"output_digest,omitempty"`
	EvidenceDigest string   `json:"evidence_digest,omitempty"`
	PartialOutput  bool     `json:"partial_output"`
	Unknown        *Unknown `json:"unknown,omitempty"`
}

type Authority struct {
	RepositoryWrites       int `json:"repository_writes"`
	BranchSettingWrites    int `json:"branch_setting_writes"`
	LocalTestExecutions    int `json:"local_test_executions"`
	CrossProjectGates      int `json:"cross_project_required_gates"`
	PromotionAuthorized    bool `json:"promotion_authorized"`
}

type Report struct {
	Schema             string             `json:"schema"`
	ProfileID          string             `json:"profile_id"`
	ProfileVersion     int                `json:"profile_version"`
	ProfileDigest      string             `json:"profile_digest"`
	Decision           string             `json:"decision"`
	Resolution         string             `json:"resolution"`
	Reason             string             `json:"reason"`
	Cells              []CellObservation  `json:"cells"`
	Summary            Summary            `json:"summary"`
	Counterexamples    []Counterexample   `json:"counterexamples"`
	Activities         []string           `json:"activities"`
	ActivityCount      int                `json:"activity_count"`
	BindingCount       int                `json:"binding_count"`
	EvidenceDigests    map[string]string  `json:"evidence_digests"`
	Authority          Authority          `json:"authority"`
	Improvement        string             `json:"improvement"`
	HumanReport        string             `json:"human_report"`
}
