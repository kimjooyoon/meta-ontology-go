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
	ObjectFormat        string                   `json:"object_format"`
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
	Path                string `json:"path"`
	Tracked             bool   `json:"tracked"`
	Kind                string `json:"kind"`
	Mode                string `json:"mode"`
	SymlinkTargetDigest string `json:"symlink_target_digest,omitempty"`
	ContentDigest       string `json:"content_digest"`
	ObjectFormat        string `json:"object_format"`
	ObjectID            string `json:"object_id"`
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
	Path                 string `json:"path"`
	BindingMode          string `json:"binding_mode"`
	ObservedCheckoutSHA  string `json:"observed_checkout_sha"`
	HeadPathObjectID     string `json:"head_path_object_id"`
	ObjectFormat         string `json:"object_format"`
	ActualSourceObjectID string `json:"actual_source_object_id"`
	SourceBytesDigest    string `json:"source_bytes_digest"`
	RawDigest            string `json:"raw_digest"`
	ParsedDigest         string `json:"parsed_digest"`
	SemanticDigest       string `json:"semantic_digest"`
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
	Stage            string   `json:"stage"`
	Step             string   `json:"step"`
	Reason           string   `json:"reason"`
	SubjectPath      string   `json:"subject_path,omitempty"`
	ClaimInstanceIDs []string `json:"claim_instance_ids,omitempty"`
	Edges            []string `json:"edges"`
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

type PlanGate struct {
	Decision              string `json:"decision"`
	Observed              int    `json:"observed"`
	Denominator           int    `json:"denominator"`
	InventoryExact        bool   `json:"inventory_exact"`
	SubjectCoverageExact  bool   `json:"subject_coverage_exact"`
	IndicatorsValid       bool   `json:"indicators_valid"`
	ClaimTransitionsValid bool   `json:"claim_transitions_valid"`
}

type Receipt struct {
	Schema               string                `json:"schema"`
	Scope                string                `json:"scope"`
	Source               SourceEvidence        `json:"source"`
	ObservationDigest    string                `json:"observation_digest"`
	Operation            Operation             `json:"operation"`
	ExecutionMode        string                `json:"execution_mode"`
	Execution            ExecutionStatus       `json:"execution"`
	Conformance          Conformance           `json:"conformance"`
	PlanGate             PlanGate              `json:"plan_gate"`
	PolicyContradictions []PolicyContradiction `json:"policy_contradictions"`
	Subjects             []SubjectResolution   `json:"subjects"`
	ClaimTransitions     []ClaimTransition     `json:"claim_transitions"`
	Metrics              Metrics               `json:"metrics"`
	CheckInventory       ExactInventory        `json:"check_inventory"`
	IndicatorInventory   ExactInventory        `json:"indicator_inventory"`
	Indicators           []Indicator           `json:"indicators"`
	IndependentVerifier  IndependentVerifier   `json:"independent_verifier"`
	PlanDigest           string                `json:"plan_digest"`
	Digest               string                `json:"digest"`
}

type InterventionResult struct {
	ID                   string                `json:"id"`
	Source               SourceEvidence        `json:"source"`
	Conformance          Conformance           `json:"conformance"`
	PlanGate             PlanGate              `json:"plan_gate"`
	PolicyContradictions []PolicyContradiction `json:"policy_contradictions"`
	Subjects             []SubjectResolution   `json:"subjects"`
	ClaimTransitions     []ClaimTransition     `json:"claim_transitions"`
	Execution            ExecutionStatus       `json:"execution"`
	PlanDigest           string                `json:"plan_digest"`
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
	LogicalSourcePath string     `json:"logical_source_path"`
	BindingMode       string     `json:"binding_mode"`
	PlanReceiptDigest string     `json:"plan_receipt_digest"`
	InputDigest       string     `json:"input_digest"`
	SourceBytesDigest string     `json:"source_bytes_digest"`
	ResultDigest      string     `json:"result_digest"`
	ConsumerIdentity  string     `json:"consumer_identity"`
	ExitCode          int        `json:"exit_code"`
	Result            string     `json:"result"`
	Coordinate        Coordinate `json:"coordinate"`
	Digest            string     `json:"digest"`
}

type ProcessExecutionEvidence struct {
	VariantID            string `json:"variant_id"`
	SelfReportedExitCode int    `json:"self_reported_exit_code"`
	SelfReportedResult   string `json:"self_reported_result"`
	ObservedOSExitCode   int    `json:"observed_os_exit_code"`
	ObservedStdoutDigest string `json:"observed_stdout_digest"`
	ObservedStdoutBytes  int    `json:"observed_stdout_bytes"`
	ObservedStderrDigest string `json:"observed_stderr_digest"`
	ObservedStderrBytes  int    `json:"observed_stderr_bytes"`
	ResultDigest         string `json:"result_digest"`
}

type SourcePlanBindingEvidence struct {
	VariantID                       string `json:"variant_id"`
	PlanReceiptDigest               string `json:"plan_receipt_digest"`
	ExpectedSourceRawDigest         string `json:"expected_source_raw_digest"`
	ExpectedSourceBytesDigest       string `json:"expected_source_bytes_digest"`
	ExpectedSourceObjectID          string `json:"expected_source_object_id"`
	ExpectedObjectFormat            string `json:"expected_object_format"`
	ActualConsumerSourceBytesDigest string `json:"actual_consumer_source_bytes_digest"`
	ActualConsumerSourceObjectID    string `json:"actual_consumer_source_object_id"`
	LogicalSourcePath               string `json:"logical_source_path"`
	BindingMode                     string `json:"binding_mode"`
	ExpectedBindingMode             string `json:"expected_binding_mode"`
	Exact                           bool   `json:"exact"`
}

type GoRuntimeEvidence struct {
	ExpectedVersion            string             `json:"expected_version"`
	GoVersion                  string             `json:"go_version"`
	GoEnvGOVERSION             string             `json:"go_env_goversion"`
	FixerInventory             []string           `json:"fixer_inventory"`
	FixHelpDigest              string             `json:"fix_help_digest"`
	FixHelpStderrDigest        string             `json:"fix_help_stderr_digest"`
	FixHelpStderrBytes         int                `json:"fix_help_stderr_bytes"`
	FixDiffStdoutDigest        string             `json:"fix_diff_stdout_digest"`
	FixDiffStderrDigest        string             `json:"fix_diff_stderr_digest"`
	FixDiffStdoutBytes         int                `json:"fix_diff_stdout_bytes"`
	FixDiffStderrBytes         int                `json:"fix_diff_stderr_bytes"`
	FixDiffStderrAllowed       bool               `json:"fix_diff_stderr_allowed"`
	FixDiffExitCode            int                `json:"fix_diff_exit_code"`
	FixHelpCommandArgv         []string           `json:"fix_help_command_argv"`
	FixDiffCommandArgv         []string           `json:"fix_diff_command_argv"`
	CommandCWD                 string             `json:"command_cwd"`
	SubjectHeadSHA             string             `json:"subject_head_sha"`
	SubjectTreeID              string             `json:"subject_tree_id"`
	ObjectFormat               string             `json:"object_format"`
	GoModDigest                string             `json:"go_mod_digest"`
	GoSumDigest                string             `json:"go_sum_digest"`
	PackageUniverseCount       int                `json:"package_universe_count"`
	PackageUniverseNumerator   int                `json:"package_universe_numerator"`
	PackageUniverseDenominator int                `json:"package_universe_denominator"`
	PackageUniverseDigest      string             `json:"package_universe_digest"`
	PackageListCommandArgv     []string           `json:"package_list_command_argv"`
	RequiredFixers             []string           `json:"required_fixers"`
	RequiredFixersSatisfied    int                `json:"required_fixers_satisfied"`
	RequiredFixersDenominator  int                `json:"required_fixers_denominator"`
	RemovedFixerID             string             `json:"removed_fixer_id"`
	RemovedFixerPresent        bool               `json:"removed_fixer_present"`
	RemovedFixerNumerator      int                `json:"removed_fixer_numerator"`
	RemovedFixerDenominator    int                `json:"removed_fixer_denominator"`
	ActiveFixFixture           FixFixtureEvidence `json:"active_fix_fixture"`
	Conformant                 bool               `json:"conformant"`
}

type FixFixtureEvidence struct {
	CommandArgv  []string `json:"command_argv"`
	ExitCode     int      `json:"exit_code"`
	StdoutDigest string   `json:"stdout_digest"`
	StdoutBytes  int      `json:"stdout_bytes"`
	StderrDigest string   `json:"stderr_digest"`
	StderrBytes  int      `json:"stderr_bytes"`
	ExpectedDiff bool     `json:"expected_diff"`
	ObservedDiff bool     `json:"observed_diff"`
	Numerator    int      `json:"numerator"`
	Denominator  int      `json:"denominator"`
	Conformant   bool     `json:"conformant"`
}

type AdjudicationReceipt struct {
	Schema                       string                      `json:"schema"`
	Scope                        string                      `json:"scope"`
	ObservationDigest            string                      `json:"observation_digest"`
	ExpectedVariantIDs           []string                    `json:"expected_variant_ids"`
	ObservedVariantIDs           []string                    `json:"observed_variant_ids"`
	Adjudications                []ConsumerAdjudication      `json:"adjudications"`
	ProcessEvidence              []ProcessExecutionEvidence  `json:"process_evidence"`
	SourcePlanBinding            []SourcePlanBindingEvidence `json:"source_plan_binding"`
	SourcePlanBindingNumer       int                         `json:"source_plan_binding_numerator"`
	SourcePlanBindingDenom       int                         `json:"source_plan_binding_denominator"`
	SourceReconstructionNumer    int                         `json:"source_reconstruction_numerator"`
	SourceReconstructionDenom    int                         `json:"source_reconstruction_denominator"`
	Decision                     string                      `json:"decision"`
	Coordinate                   Coordinate                  `json:"coordinate"`
	SelectedCheckExecutionSchema string                      `json:"selected_check_execution_schema"`
	SelectedCheckExecution       string                      `json:"selected_check_execution"`
	SelectedCheckExecutionNumer  int                         `json:"selected_check_execution_numerator"`
	SelectedCheckExecutionDenom  int                         `json:"selected_check_execution_denominator"`
	GoRuntime                    GoRuntimeEvidence           `json:"go_runtime"`
	CIEvidence                   CIEvidenceAdjudication      `json:"ci_evidence"`
	Digest                       string                      `json:"digest"`
}
