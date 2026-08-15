package bindingcoverage

const SchemaVersion = "gooo/selective-ci-binding-coverage/v1"

type BindingKind string

const (
	KindExactValue    BindingKind = "EXACT_VALUE"
	KindExactDigest   BindingKind = "EXACT_DIGEST"
	KindSetEqual      BindingKind = "SET_EQUAL"
	KindDerivedDigest BindingKind = "DERIVED_DIGEST"
	ExactValue                    = KindExactValue
	ExactDigest                   = KindExactDigest
	SetEqual                      = KindSetEqual
	DerivedDigest                 = KindDerivedDigest
)

type Polarity string

const (
	PolarityMatch    Polarity = "MATCH"
	PolarityMismatch Polarity = "MISMATCH"
	MATCH                     = PolarityMatch
	MISMATCH                  = PolarityMismatch
)

type Decision string

const (
	DecisionExact      Decision = "EXACT"
	DecisionIncomplete Decision = "INCOMPLETE"
	DecisionUnknown    Decision = "UNKNOWN"
	EXACT                       = DecisionExact
	INCOMPLETE                  = DecisionIncomplete
	UNKNOWN                     = DecisionUnknown
)

type Reason string

const (
	ReasonComplete                Reason = "COMPLETE"
	ReasonZeroDenominator         Reason = "ZERO_DENOMINATOR"
	ReasonMissingMatch            Reason = "MISSING_MATCH"
	ReasonMissingMismatch         Reason = "MISSING_MISMATCH"
	ReasonMissingMatchAndMismatch Reason = "MISSING_MATCH_AND_MISMATCH"
	ReasonUnknownSchema           Reason = "UNKNOWN_SCHEMA"
	ReasonMissingInput            Reason = "MISSING_INPUT"
	ReasonInvalidID               Reason = "INVALID_ID"
	ReasonInvalidDigest           Reason = "INVALID_DIGEST"
	ReasonInvalidEnum             Reason = "INVALID_ENUM"
	ReasonInvalidToken            Reason = "INVALID_TOKEN"
	ReasonDuplicateID             Reason = "DUPLICATE_ID"
	ReasonUnknownReference        Reason = "UNKNOWN_REFERENCE"
	ReasonDuplicatePolarity       Reason = "DUPLICATE_POLARITY"
	ReasonWorkOverflow            Reason = "WORK_OVERFLOW"
	ReasonEvaluatorError          Reason = "EVALUATOR_ERROR"
	ReasonStaleInput              Reason = "STALE_INPUT"
	ReasonSnapshotMismatch        Reason = "SNAPSHOT_MISMATCH"
	ReasonInvalidPrecedence       Reason = "INVALID_PRECEDENCE"
	ReasonDuplicatePrecedence     Reason = "DUPLICATE_PRECEDENCE"
	ReasonUnregisteredPair        Reason = "UNREGISTERED_PRECEDENCE"
	ReasonStalePartition          Reason = "STALE_PARTITION"
	ReasonSelfLink                Reason = "SELF_LINK"
)

type RequiredBinding struct {
	BindingID      string      `json:"binding_id"`
	FromFieldID    string      `json:"from_field_id"`
	ToFieldID      string      `json:"to_field_id"`
	Kind           BindingKind `json:"kind"`
	ExpectedStage  string      `json:"expected_stage"`
	ExpectedReason string      `json:"expected_reason"`
}

type PrecedenceEntry struct {
	Rank   uint64 `json:"rank"`
	Stage  string `json:"stage"`
	Reason string `json:"reason"`
}

type Partition struct {
	PartitionID    string   `json:"partition_id"`
	BindingID      string   `json:"binding_id"`
	Polarity       Polarity `json:"polarity"`
	ExpectedStage  string   `json:"expected_stage"`
	ExpectedReason string   `json:"expected_reason"`
}

type Input struct {
	SchemaVersion          string            `json:"schema_version"`
	ContractID             string            `json:"contract_id"`
	SnapshotDigest         string            `json:"snapshot_digest"`
	ExpectedSnapshotDigest string            `json:"expected_snapshot_digest"`
	RequiredBindings       []RequiredBinding `json:"required_bindings"`
	Partitions             []Partition       `json:"partitions"`
	PrecedenceRegistry     []PrecedenceEntry `json:"precedence_registry"`
}

type Output struct {
	SchemaVersion             string   `json:"schema_version"`
	ContractID                string   `json:"contract_id"`
	SnapshotDigest            string   `json:"snapshot_digest"`
	ExpectedSnapshotDigest    string   `json:"expected_snapshot_digest"`
	InputDigest               string   `json:"input_digest"`
	RequiredBindingCount      uint64   `json:"required_binding_count"`
	MatchCoveredCount         uint64   `json:"match_covered_count"`
	MismatchCoveredCount      uint64   `json:"mismatch_covered_count"`
	PartitionCount            uint64   `json:"partition_count"`
	EndpointReferenceCount    uint64   `json:"endpoint_reference_count"`
	DeterministicWorkUnits    uint64   `json:"deterministic_work_units"`
	InputBytes                uint64   `json:"input_bytes"`
	Decision                  Decision `json:"decision"`
	Reason                    Reason   `json:"reason"`
	MissingMatchBindingIDs    []string `json:"missing_match_binding_ids"`
	MissingMismatchBindingIDs []string `json:"missing_mismatch_binding_ids"`
	CanonicalDigest           string   `json:"canonical_digest"`
}
