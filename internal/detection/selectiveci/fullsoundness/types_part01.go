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
