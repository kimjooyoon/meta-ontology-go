package causalci

// Observation is raw CI evidence. It contains facts collected from the PR
// checkout and deliberately contains no decision, choice, or reason fields.
type Observation struct {
	Schema              string                   `json:"schema"`
	Repository          string                   `json:"repository"`
	BaseSHA             string                   `json:"base_sha"`
	HeadSHA             string                   `json:"head_sha"`
	ObservedCheckoutSHA string                   `json:"observed_checkout_sha"`
	SourcePath          string                   `json:"source_path"`
	HeadPathObjectID    string                   `json:"head_path_object_id"`
	SourceBytesDigest   string                   `json:"source_bytes_digest"`
	ChangedFiles        []ChangedFileObservation `json:"changed_files"`
	PriorClaims         []PriorClaimObservation  `json:"prior_claims"`
	Isolation           IsolationObservation     `json:"isolation"`
}

type ChangedFileObservation struct {
	Path         string `json:"path"`
	Status       string `json:"status"`
	BeforeObject string `json:"before_object,omitempty"`
	AfterObject  string `json:"after_object,omitempty"`
}

// Prior claims are content-addressed instances of a raw claim template.
type PriorClaimObservation struct {
	TemplateID  string `json:"template_id"`
	InstanceID  string `json:"instance_id"`
	SubjectPath string `json:"subject_path"`
	Proposition string `json:"proposition"`
	State       string `json:"state"`
	Provenance  string `json:"provenance"`
}

type IsolationObservation struct {
	Before RepositorySnapshot `json:"before"`
	After  RepositorySnapshot `json:"after"`
}

// Entries include tracked and untracked paths plus content digests. A status
// line set is insufficient to establish that a process did not alter data.
type RepositorySnapshot struct {
	Entries        []RepositoryEntry `json:"entries"`
	SnapshotDigest string            `json:"snapshot_digest"`
}

type RepositoryEntry struct {
	Path          string `json:"path"`
	Tracked       bool   `json:"tracked"`
	ContentDigest string `json:"content_digest"`
}

// PolicyGraph is generated from parsed, lowered, semantic Gooo source.
type PolicyGraph struct {
	Source          SourceEvidence
	ChangedFileID   string
	ClaimID         string
	SurfaceID       string
	Checks          []Check
	Edges           []PolicyEdge
	PriorStates     []PriorStateRule
	Contradictions  []PolicyContradiction
	ClaimStateRules map[string]string
}

type SourceEvidence struct {
	Path                string `json:"path"`
	BindingMode         string `json:"binding_mode"`
	ObservedCheckoutSHA string `json:"observed_checkout_sha"`
	HeadPathObjectID    string `json:"head_path_object_id"`
	SourceBytesDigest   string `json:"source_bytes_digest"`
	RawDigest           string `json:"raw_digest"`
	ParsedDigest        string `json:"parsed_digest"`
	SemanticDigest      string `json:"semantic_digest"`
}

type Check struct {
	ID         string `json:"id"`
	Ordinal    int    `json:"ordinal"`
	SemanticID string `json:"semantic_id"`
}

type PolicyEdge struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	From         string `json:"from"`
	To           string `json:"to"`
	ActivityID   string `json:"activity_id"`
	ValueProgram string `json:"value_program"`
}

type PriorStateRule struct {
	State        string `json:"state"`
	ActivityID   string `json:"activity_id"`
	ValueProgram string `json:"value_program"`
}

type PolicyContradiction struct {
	Stage  string   `json:"stage"`
	Step   string   `json:"step"`
	Reason string   `json:"reason"`
	Edges  []string `json:"edges"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Operation struct {
	Producer                string                     `json:"producer"`
	Consumer                string                     `json:"consumer"`
	MetaOperation           string                     `json:"meta_operation"`
	ProofChoice             string                     `json:"proof_choice"`
	DeclaredPlanCapability  string                     `json:"declared_plan_capability"`
	ObservedRepositoryState RepositoryStateObservation `json:"observed_repository_state"`
}

type RepositoryStateObservation struct {
	NetState                string `json:"net_state"`
	ChangedPathCount        int    `json:"changed_path_count"`
	ChangedContentCount     int    `json:"changed_content_count"`
	TransientWrites         string `json:"transient_writes"`
	GlobalMutationAuthority string `json:"global_mutation_authority"`
}

type PathEvidence struct {
	SubjectPath    string   `json:"subject_path"`
	ClaimIDs       []string `json:"claim_ids"`
	Proposition    string   `json:"proposition"`
	SurfaceID      string   `json:"surface_id"`
	CheckID        string   `json:"check_id"`
	PolicyEdgeIDs  []string `json:"policy_edge_ids"`
	SemanticDigest string   `json:"semantic_digest"`
	Explanation    string   `json:"explanation"`
	ProofChoice    string   `json:"proof_choice"`
}

type UnknownCause struct {
	SubjectPath string     `json:"subject_path"`
	Coordinate  Coordinate `json:"coordinate"`
	Provenance  string     `json:"provenance"`
}

type CheckChoice struct {
	CheckID     string   `json:"check_id"`
	ProofChoice string   `json:"proof_choice"`
	Reason      string   `json:"reason"`
	ClaimIDs    []string `json:"claim_ids,omitempty"`
	PathIDs     []string `json:"path_ids,omitempty"`
}

type SubjectResolution struct {
	Path           string         `json:"path"`
	Resolution     string         `json:"resolution"`
	Coordinate     Coordinate     `json:"coordinate"`
	Paths          []PathEvidence `json:"paths,omitempty"`
	UnknownCauses  []UnknownCause `json:"unknown_causes,omitempty"`
	SelectedChecks []CheckChoice  `json:"selected_checks"`
}

type ClaimTransition struct {
	Sequence       int    `json:"sequence"`
	TemplateID     string `json:"template_id"`
	ClaimID        string `json:"claim_id"`
	SubjectPath    string `json:"subject_path"`
	Proposition    string `json:"proposition"`
	Before         string `json:"before"`
	After          string `json:"after"`
	Resolution     string `json:"resolution"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	PreviousDigest string `json:"previous_digest"`
	Digest         string `json:"digest"`
}

type Conformance struct {
	Decision   string     `json:"decision"`
	Coordinate Coordinate `json:"coordinate"`
}

type ExactInventory struct {
	ExpectedIDs []string `json:"expected_ids"`
	ObservedIDs []string `json:"observed_ids"`
}

type Metrics struct {
	SubjectUniverseDigest      string `json:"subject_universe_digest"`
	SubjectUniverseCount       int    `json:"subject_universe_count"`
	SubjectCoverageNumerator   int    `json:"subject_coverage_numerator"`
	SubjectCoverageDenominator int    `json:"subject_coverage_denominator"`
	SubjectTotal               int    `json:"subject_total"`
	SelectedSubjectTotal       int    `json:"selected_subject_total"`
	UnknownSubjectTotal        int    `json:"unknown_subject_total"`
	FailClosedSubjectTotal     int    `json:"fail_closed_subject_total"`
	SelectedCheckTotal         int    `json:"selected_check_total"`
	FullSuiteCheckDenominator  int    `json:"full_suite_check_denominator"`
	ClaimTransitionTotal       int    `json:"claim_transition_total"`
	DischargedClaimTotal       int    `json:"discharged_claim_total"`
	LowerResolutionClaimTotal  int    `json:"lower_resolution_claim_total"`
	RefutedClaimTotal          int    `json:"refuted_claim_total"`
	FixedIndicatorSatisfied    int    `json:"fixed_indicator_satisfied"`
	FixedIndicatorDenominator  int    `json:"fixed_indicator_denominator"`
}

type Indicator struct {
	ID          string `json:"id"`
	Observed    int    `json:"observed"`
	Denominator int    `json:"denominator"`
	Satisfied   bool   `json:"satisfied"`
}

type IndependentVerifier struct {
	ID         string `json:"id"`
	Mode       string `json:"mode"`
	Required   bool   `json:"required"`
	Capability string `json:"capability"`
}

type ExecutionStatus struct {
	Result     string     `json:"result"`
	Capability string     `json:"capability"`
	Coordinate Coordinate `json:"coordinate"`
}

type Receipt struct {
	Schema              string              `json:"schema"`
	Scope               string              `json:"scope"`
	Source              SourceEvidence      `json:"source"`
	ObservationDigest   string              `json:"observation_digest"`
	Operation           Operation           `json:"operation"`
	ExecutionMode       string              `json:"execution_mode"`
	Execution           ExecutionStatus     `json:"execution"`
	Conformance         Conformance         `json:"conformance"`
	Subjects            []SubjectResolution `json:"subjects"`
	ClaimTransitions    []ClaimTransition   `json:"claim_transitions"`
	Metrics             Metrics             `json:"metrics"`
	CheckInventory      ExactInventory      `json:"check_inventory"`
	IndicatorInventory  ExactInventory      `json:"indicator_inventory"`
	Indicators          []Indicator         `json:"indicators"`
	IndependentVerifier IndependentVerifier `json:"independent_verifier"`
	PlanDigest          string              `json:"plan_digest"`
	Digest              string              `json:"digest"`
}

type InterventionResult struct {
	ID               string              `json:"id"`
	Source           SourceEvidence      `json:"source"`
	Conformance      Conformance         `json:"conformance"`
	Subjects         []SubjectResolution `json:"subjects"`
	ClaimTransitions []ClaimTransition   `json:"claim_transitions"`
	Execution        ExecutionStatus     `json:"execution"`
	PlanDigest       string              `json:"plan_digest"`
}

type InterventionReport struct {
	Schema             string             `json:"schema"`
	ObservationDigest  string             `json:"observation_digest"`
	ExpectedVariantIDs []string           `json:"expected_variant_ids"`
	ObservedVariantIDs []string           `json:"observed_variant_ids"`
	Base               InterventionResult `json:"base"`
	Semantic           InterventionResult `json:"semantic"`
	Nonsemantic        InterventionResult `json:"nonsemantic"`
	Contradiction      InterventionResult `json:"contradiction"`
	Digest             string             `json:"digest"`
}

// ConsumerAdjudication is written by the independent consumer process after
// it has actually replayed one producer receipt.
type ConsumerAdjudication struct {
	Schema            string     `json:"schema"`
	VariantID         string     `json:"variant_id"`
	PlanReceiptDigest string     `json:"plan_receipt_digest"`
	InputDigest       string     `json:"input_digest"`
	SourceBytesDigest string     `json:"source_bytes_digest"`
	OutputDigest      string     `json:"output_digest"`
	ConsumerIdentity  string     `json:"consumer_identity"`
	ExitCode          int        `json:"exit_code"`
	Result            string     `json:"result"`
	Coordinate        Coordinate `json:"coordinate"`
	Digest            string     `json:"digest"`
}

type GoRuntimeEvidence struct {
	ExpectedVersion     string   `json:"expected_version"`
	GoVersion           string   `json:"go_version"`
	GoEnvGOVERSION      string   `json:"go_env_goversion"`
	FixerInventory      []string `json:"fixer_inventory"`
	FixHelpDigest       string   `json:"fix_help_digest"`
	FixHelpStderrDigest string   `json:"fix_help_stderr_digest"`
	FixHelpStderrBytes  int      `json:"fix_help_stderr_bytes"`
	FixDiffStdoutDigest string   `json:"fix_diff_stdout_digest"`
	FixDiffStderrDigest string   `json:"fix_diff_stderr_digest"`
	FixDiffStdoutBytes  int      `json:"fix_diff_stdout_bytes"`
	FixDiffStderrBytes  int      `json:"fix_diff_stderr_bytes"`
	FixDiffExitCode     int      `json:"fix_diff_exit_code"`
	Conformant          bool     `json:"conformant"`
}

type AdjudicationReceipt struct {
	Schema                    string                 `json:"schema"`
	Scope                     string                 `json:"scope"`
	ObservationDigest         string                 `json:"observation_digest"`
	ExpectedVariantIDs        []string               `json:"expected_variant_ids"`
	ObservedVariantIDs        []string               `json:"observed_variant_ids"`
	Adjudications             []ConsumerAdjudication `json:"adjudications"`
	SourceReconstructionNumer int                    `json:"source_reconstruction_numerator"`
	SourceReconstructionDenom int                    `json:"source_reconstruction_denominator"`
	Decision                  string                 `json:"decision"`
	Coordinate                Coordinate             `json:"coordinate"`
	GoRuntime                 GoRuntimeEvidence      `json:"go_runtime"`
	Digest                    string                 `json:"digest"`
}
