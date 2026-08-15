package fullsoundness

const SchemaVersion = "gooo/selective-ci-full-soundness/v1"

type Decision string

const (
	DecisionSound   Decision = "SOUND"
	DecisionUnsound Decision = "UNSOUND"
	DecisionUnknown Decision = "UNKNOWN"
)

type Reason string

const (
	ReasonSound                      Reason = "SOUND"
	ReasonAuthorizationPresent       Reason = "AUTHORIZATION_PRESENT"
	ReasonSelectedSetMismatch        Reason = "SELECTED_SET_MISMATCH"
	ReasonGlobalGuardOmitted         Reason = "GLOBAL_GUARD_OMITTED"
	ReasonSelectedExtraFailure       Reason = "SELECTED_EXTRA_FAILURE"
	ReasonSelectedFullStatusMismatch Reason = "SELECTED_FULL_STATUS_MISMATCH"
	ReasonFailureCodeMismatch        Reason = "FAILURE_CODE_MISMATCH"
	ReasonOutputDigestMismatch       Reason = "OUTPUT_DIGEST_MISMATCH"
	ReasonOmittedFullFailure         Reason = "OMITTED_FULL_FAILURE"
	ReasonImpactedCommandOmitted     Reason = "IMPACTED_COMMAND_OMITTED"
	ReasonFullSuiteRequired          Reason = "FULL_SUITE_REQUIRED"
	ReasonUnregisteredObligation     Reason = "UNREGISTERED_OBLIGATION"
	ReasonUnprovableObligation       Reason = "UNPROVABLE_OBLIGATION"
	ReasonZeroCommandDenominator     Reason = "ZERO_COMMAND_DENOMINATOR"
	ReasonInvalidOutcome             Reason = "INVALID_OUTCOME"
	ReasonDigestBindingMismatch      Reason = "DIGEST_BINDING_MISMATCH"
	ReasonResourceOverflow           Reason = "RESOURCE_OVERFLOW"
)

type ObligationAuthority string

const (
	AuthorityAuthoritative ObligationAuthority = "AUTHORITATIVE"
	AuthorityCandidate     ObligationAuthority = "CANDIDATE"
	AuthorityDerived       ObligationAuthority = "DERIVED"
)

type OutcomeStatus string

const (
	OutcomePass OutcomeStatus = "PASS"
	OutcomeFail OutcomeStatus = "FAIL"
)

type ResourceClass string

const (
	ResourceImproved    ResourceClass = "IMPROVED"
	ResourceEqual       ResourceClass = "EQUAL"
	ResourceRegressed   ResourceClass = "REGRESSED"
	ResourceNotComputed ResourceClass = "NOT_COMPUTED"
)

type Obligation struct {
	ID        string              `json:"id"`
	Authority ObligationAuthority `json:"authority"`
}

type Command struct {
	ID            string   `json:"id"`
	ObligationIDs []string `json:"obligation_ids"`
	GlobalGuard   bool     `json:"global_guard"`
}

type SelectionReceipt struct {
	SnapshotDigest  string   `json:"snapshot_digest"`
	PolicyDigest    string   `json:"policy_digest"`
	RegistryDigest  string   `json:"registry_digest"`
	SelectionDigest string   `json:"selection_digest"`
	CommandIDs      []string `json:"command_ids"`
}

type Outcome struct {
	CommandID    string        `json:"command_id"`
	Status       OutcomeStatus `json:"status"`
	FailureCode  string        `json:"failure_code"`
	OutputDigest string        `json:"output_digest"`
}

type ResourceReceipt struct {
	CommandID         string `json:"command_id"`
	SnapshotDigest    string `json:"snapshot_digest"`
	ToolchainDigest   string `json:"toolchain_digest"`
	RunnerDigest      string `json:"runner_digest"`
	CPUCoreNS         int64  `json:"cpu_core_ns"`
	AllocatedCPUCount int64  `json:"allocated_cpu_count"`
	WallNS            int64  `json:"wall_ns"`
	PeakRSSBytes      int64  `json:"peak_rss_bytes"`
	ReadBytes         int64  `json:"read_bytes"`
	WriteBytes        int64  `json:"write_bytes"`
}

type Input struct {
	SchemaVersion            string            `json:"schema_version"`
	SnapshotDigest           string            `json:"snapshot_digest"`
	PolicyDigest             string            `json:"policy_digest"`
	RegistryDigest           string            `json:"registry_digest"`
	SelectionDigest          string            `json:"selection_digest"`
	ToolchainDigest          string            `json:"toolchain_digest"`
	RunnerDigest             string            `json:"runner_digest"`
	Obligations              []Obligation      `json:"obligations"`
	Commands                 []Command         `json:"commands"`
	ImpactedObligationIDs    []string          `json:"impacted_obligation_ids"`
	SelectedCommandIDs       []string          `json:"selected_command_ids"`
	SelectionReceipt         *SelectionReceipt `json:"selection_receipt"`
	FullOutcomes             []Outcome         `json:"full_outcomes"`
	SelectedOutcomes         []Outcome         `json:"selected_outcomes"`
	FullResourceReceipts     []ResourceReceipt `json:"full_resource_receipts"`
	SelectedResourceReceipts []ResourceReceipt `json:"selected_resource_receipts"`
	ExecutionAuthorized      bool              `json:"execution_authorized"`
	CIAuthorized             bool              `json:"ci_authorized"`
}

type Utilization struct {
	Numerator   int64 `json:"numerator"`
	Denominator int64 `json:"denominator"`
}

type ResourceTotals struct {
	CPUCoreNS    int64       `json:"cpu_core_ns"`
	PeakRSSBytes int64       `json:"peak_rss_bytes"`
	ReadBytes    int64       `json:"read_bytes"`
	WriteBytes   int64       `json:"write_bytes"`
	Utilization  Utilization `json:"utilization"`
}

type ResourceVector struct {
	Full     ResourceTotals `json:"full"`
	Selected ResourceTotals `json:"selected"`
	Class    ResourceClass  `json:"class"`
}

type Output struct {
	SchemaVersion                        string          `json:"schema_version"`
	SnapshotDigest                       string          `json:"snapshot_digest"`
	PolicyDigest                         string          `json:"policy_digest"`
	RegistryDigest                       string          `json:"registry_digest"`
	SelectionDigest                      string          `json:"selection_digest"`
	CommandCount                         uint64          `json:"command_count"`
	SelectedCommandCount                 uint64          `json:"selected_command_count"`
	ObligationCount                      uint64          `json:"obligation_count"`
	AuthoritativeImpactedObligationCount uint64          `json:"authoritative_impacted_obligation_count"`
	SemanticEvaluated                    bool            `json:"semantic_evaluated"`
	FullFailureCommandIDs                []string        `json:"full_failure_command_ids"`
	SelectedFailureCommandIDs            []string        `json:"selected_failure_command_ids"`
	OmittedCommandIDs                    []string        `json:"omitted_command_ids"`
	ResourceVector                       *ResourceVector `json:"resource_vector"`
	Decision                             Decision        `json:"decision"`
	Reason                               Reason          `json:"reason"`
	ExecutionAuthorized                  bool            `json:"execution_authorized"`
	CIAuthorized                         bool            `json:"ci_authorized"`
	DecisionDigest                       string          `json:"decision_digest"`
	CanonicalDigest                      string          `json:"canonical_digest"`
}
