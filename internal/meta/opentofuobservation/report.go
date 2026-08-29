package opentofuobservation

type CellResult struct {
	ID             string `json:"id"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	Indicator      string `json:"indicator"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	Decision       string `json:"decision"`
	State          string `json:"state"`
	Observed       int    `json:"observed"`
	Expected       int    `json:"expected"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Summary struct {
	CellsTotal       int `json:"cells_total"`
	ClosedCells      int `json:"closed_cells"`
	OpenCells        int `json:"open_cells"`
	UnknownCells     int `json:"unknown_cells"`
	RefutedCells     int `json:"refuted_cells"`
	FoundationClosed int `json:"foundation_closed"`
	CoherenceClosed  int `json:"coherence_closed"`
	RegressionClosed int `json:"regression_closed"`
	ThreePaths       int `json:"three_user_paths"`
	Executions       int `json:"executions"`
	ReplayMatches    int `json:"replay_matches"`
	RepositoryWrites int `json:"repository_writes"`
	LocalTests       int `json:"local_test_executions"`
}

type Report struct {
	Schema                   string             `json:"schema"`
	ContractID               string             `json:"contract_id"`
	SubjectSHA               string             `json:"subject_sha"`
	Decision                 string             `json:"decision"`
	Resolution               string             `json:"resolution"`
	Reason                   string             `json:"reason"`
	MetaOperation            string             `json:"meta_operation"`
	UserPaths                []string           `json:"user_paths"`
	Summary                  Summary            `json:"summary"`
	Cells                    []CellResult       `json:"cells"`
	Unknowns                 []Unknown          `json:"unknowns"`
	Counterexamples          []Counterexample   `json:"counterexamples"`
	Release                  ReleaseObservation `json:"release"`
	FixtureDigest            string             `json:"fixture_digest"`
	FixtureFiles             []string           `json:"fixture_files"`
	FixturePhysicalLines     int                `json:"fixture_physical_lines"`
	Executions               []ExecutionRun     `json:"executions"`
	Reuse                    ReuseAccounting    `json:"reuse"`
	Runtime                  RuntimeSummary     `json:"runtime"`
	Inventory                Inventory          `json:"inventory"`
	CellEvidenceProjections  map[string]string  `json:"cell_evidence_projections"`
	CellEvidenceDigests      map[string]string  `json:"cell_evidence_digests"`
	Graph                    GraphObservation   `json:"graph"`
	RepositoryWrites         int                `json:"repository_writes"`
	LocalTestExecutions      int                `json:"local_test_executions"`
	ReleaseBinaryBuilds      int                `json:"release_binary_builds"`
	ReleaseBinaryBuildReason string             `json:"release_binary_build_reason"`
	ObserverGoVersion        string             `json:"observer_go_version"`
	ObserverGOVERSION        string             `json:"observer_go_env_goversion"`
	ObserverToolchainDigest  string             `json:"observer_toolchain_digest"`
	GraphValidation          *GraphValidationDiagnostic `json:"graph_validation,omitempty"`
	HumanReportReady         bool               `json:"human_report_ready"`
	PromotionAuthorized      bool               `json:"promotion_authorized"`
	ReportDigest             string             `json:"report_digest"`
}
