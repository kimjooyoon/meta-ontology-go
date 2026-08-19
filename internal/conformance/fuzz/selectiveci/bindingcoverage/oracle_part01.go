package bindingcoverage

const (
	SchemaV1 = "binding-coverage/v1"

	DecisionExact      = "EXACT"
	DecisionIncomplete = "INCOMPLETE"
	DecisionUnknown    = "UNKNOWN"
)
const (
	KindExactValue    = "EXACT_VALUE"
	KindExactDigest   = "EXACT_DIGEST"
	KindSetEqual      = "SET_EQUAL"
	KindDerivedDigest = "DERIVED_DIGEST"

	PolarityMatch    = "MATCH"
	PolarityMismatch = "MISMATCH"
)

// Input is the complete binding-coverage observation. It contains no
// execution or CI authorization fields because those concerns are out of
// scope for this oracle.
type Input struct {
	Schema           string       `json:"schema"`
	SnapshotDigest   string       `json:"snapshot_digest"`
	ExpectedDigest   string       `json:"expected_snapshot_digest"`
	Precedence       []Precedence `json:"precedence"`
	RequiredBindings []Binding    `json:"required_bindings"`
	Partitions       []Partition  `json:"partitions"`
}
type Precedence struct {
	Rank   int64  `json:"rank"`
	Stage  string `json:"stage"`
	Reason string `json:"reason"`
}
type Binding struct {
	ID             string `json:"id"`
	FromFieldID    string `json:"from_field_id"`
	ToFieldID      string `json:"to_field_id"`
	Kind           string `json:"kind"`
	ExpectedStage  string `json:"expected_stage"`
	ExpectedReason string `json:"expected_reason"`
}
type Partition struct {
	BindingID string `json:"binding_id"`
	Polarity  string `json:"polarity"`
	Stage     string `json:"stage"`
	Reason    string `json:"reason"`
}

// Vector is the normalized semantic result. The two authorization values are
// deliberately fixed false: authorization is not part of this partition.
type Vector struct {
	Decision               string   `json:"decision"`
	Reason                 string   `json:"reason"`
	RequiredBindingCount   int64    `json:"required_binding_count"`
	PartitionCount         int64    `json:"partition_count"`
	EndpointReferenceCount int64    `json:"endpoint_reference_count"`
	InputBytes             int64    `json:"input_bytes"`
	WorkUnits              int64    `json:"work_units"`
	MissingMatch           []string `json:"missing_match"`
	MissingMismatch        []string `json:"missing_mismatch"`
	ExecutionAuthorized    bool     `json:"execution_authorized"`
	CIAuthorized           bool     `json:"ci_authorized"`
}
type Result struct {
	Vector
	CanonicalDigest string `json:"canonical_digest"`
}
