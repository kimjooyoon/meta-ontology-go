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
