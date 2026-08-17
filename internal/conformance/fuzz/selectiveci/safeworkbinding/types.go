package safeworkbinding

type LegacyWorkID string
type Digest string
type StableID string

const SafeWorkBindingSchemaV1 = "gooo/safe-work-binding/v1"

type SafeWorkBinding struct {
	Schema                 string   `json:"schema"`
	TaskID                 StableID `json:"task_id"`
	PathID                 StableID `json:"path_id"`
	ObligationID           StableID `json:"obligation_id"`
	SourceSnapshotDigest   Digest   `json:"source_snapshot_digest"`
	SemanticSnapshotDigest Digest   `json:"semantic_snapshot_digest"`
	PolicyDigest           Digest   `json:"policy_digest"`
	RegistryDigest         Digest   `json:"registry_digest"`
	ToolchainOptionsDigest Digest   `json:"toolchain_options_digest"`
	BindingDigest          Digest   `json:"binding_digest"`
}

type Decision uint8

const (
	DecisionPass       Decision = 0
	DecisionUnknown    Decision = 1
	DecisionFailClosed Decision = 2
)

type EnforcementEffect uint8

const EnforcementNoEffect EnforcementEffect = 0

type Reason uint8

const (
	ReasonNone                  Reason = 0
	ReasonRequiredInputMissing  Reason = 1
	ReasonInvalidUTF8           Reason = 2
	ReasonBOMForbidden          Reason = 3
	ReasonInvalidJSON           Reason = 4
	ReasonTrailingValue         Reason = 5
	ReasonDuplicateKey          Reason = 6
	ReasonUnknownField          Reason = 7
	ReasonNullValue             Reason = 8
	ReasonEmptyValue            Reason = 9
	ReasonInvalidSchema         Reason = 10
	ReasonInvalidStableID       Reason = 11
	ReasonInvalidDigest         Reason = 12
	ReasonBindingDigestMismatch Reason = 13
)

type ParseResult struct {
	Decision            Decision
	Reason              Reason
	Faults              []Reason
	FullSuiteRequired   bool
	ExecutionAuthorized bool
	EnforcementEffect   EnforcementEffect
	ResultDigest        Digest
	ReplayDigest        Digest
}
