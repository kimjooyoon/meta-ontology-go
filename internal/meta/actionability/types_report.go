package actionability

type OperationWitness struct {
	Operation          string `json:"operation"`
	IndicatorCount     int    `json:"indicator_count"`
	ProofChoice        string `json:"proof_choice"`
	BindingRegistry    string `json:"binding_registry"`
	MetaBound          bool   `json:"meta_bound"`
	Executable         bool   `json:"executable"`
	ExecutorRegistry   string `json:"executor_registry,omitempty"`
	Executor           string `json:"executor,omitempty"`
	Evaluator          string `json:"evaluator,omitempty"`
	Status             string `json:"status"`
}

type KPI struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Satisfied     bool   `json:"satisfied"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Activity      string `json:"activity"`
}

type Summary struct {
	ApplicableBlockingIndicators int `json:"applicable_blocking_indicators"`
	ActionableIndicators         int `json:"actionable_indicators"`
	UnactionableIndicators       int `json:"unactionable_indicators"`
	IndicatorCoverageBasisPoints int `json:"indicator_coverage_basis_points"`
	RequiredOperations           int `json:"required_operations"`
	ExecutableOperations         int `json:"executable_operations"`
	MissingOperations            int `json:"missing_operations"`
	OperationCoverageBasisPoints int `json:"operation_coverage_basis_points"`
	NonApplicableIndicators      int `json:"non_applicable_indicators"`
	ProjectRootExemptions        int `json:"project_root_exemptions"`
	UnboundIndicators            int `json:"unbound_indicators"`
}

type Report struct {
	Schema             string             `json:"schema"`
	CommitSHA          string             `json:"commit_sha"`
	Repository         string             `json:"repository"`
	MetricsDigest      string             `json:"metrics_digest"`
	BindingDigest      string             `json:"binding_digest"`
	RegistryDigest     string             `json:"registry_digest"`
	AuthorityDigest    string             `json:"authority_digest"`
	Decision           string             `json:"decision"`
	Reason             string             `json:"reason"`
	SelectedOperation  string             `json:"selected_operation,omitempty"`
	RootProofChoice    string             `json:"root_proof_choice"`
	RootMetaOperation  string             `json:"root_meta_operation"`
	RootActivity       string             `json:"root_activity"`
	ReplayProofChoice  string             `json:"replay_proof_choice"`
	ReplayMetaOperation string            `json:"replay_meta_operation"`
	ReplayActivity     string             `json:"replay_activity"`
	Summary            Summary            `json:"summary"`
	Indicators         []KPI              `json:"indicators"`
	Operations         []OperationWitness `json:"operations"`
	ReportDigest       string             `json:"report_digest"`
}
