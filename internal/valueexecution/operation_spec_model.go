package valueexecution

const (
	OperationSpecSchema = "gooo.operation-spec/v1"
	OperationIRSchema   = "gooo.operation-invocation-ir/v1"
	IntegerEntity       = "Integer"
	OperandInt64Literal = "INT64_LITERAL"
	EffectPureValue     = "PURE_VALUE"
	Deterministic       = "DETERMINISTIC"
)

type OperationAuthority struct {
	RepositoryWrite bool `json:"repository_write"`
	ExternalCall    bool `json:"external_call"`
	Promotion       bool `json:"promotion"`
}

type OperationSpec struct {
	Schema         string             `json:"schema"`
	ID             string             `json:"id"`
	Version        int                `json:"version"`
	Arity          int                `json:"arity"`
	InputEntities  []string           `json:"input_entities"`
	OperandKind    string             `json:"operand_kind"`
	OutputEntity   string             `json:"output_entity"`
	Effect         string             `json:"effect"`
	Determinism    string             `json:"determinism"`
	FailureReasons []string           `json:"failure_reasons"`
	Authority      OperationAuthority `json:"authority"`
}

type OperandIR struct {
	Kind  string `json:"kind"`
	Int64 int64  `json:"int64"`
}

type OperationIR struct {
	Schema        string        `json:"schema"`
	Activity      string        `json:"activity"`
	Program       string        `json:"program"`
	Spec          OperationSpec `json:"spec"`
	SpecDigest    string        `json:"spec_digest"`
	InputEntities []string      `json:"input_entities"`
	OutputEntity  string        `json:"output_entity"`
	Operand       OperandIR     `json:"operand"`
}
