package selfimprovementvaluewitnessinput

import "github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"

const (
	Schema           = "gooo/self-improvement-value-witness-execution-input/v1"
	ContractID       = "gooo://self-improvement/value-witness-execution-input/v1"
	SourcePath       = "examples/language-value-witness/main.gooo"
	ActivityName     = "Increment"
	OperationID      = "self-improvement.value-witness-experiment.v1"
	BoundedTarget    = "VALUE_WITNESS_EXPERIMENT"
	Phase            = "VALUE_WITNESS_DECLARATION"
	ValueProgram     = "int.add:1"
	InputAuthority   = "CALLER_OWNED_INPUT"
	OutputAuthority  = "CALLER_OWNED_OUTPUT"
	OutputSchema     = valueexecution.ReportSchema
	EvaluatorID      = "internal/valueexecution"
	EvaluatorVersion = "v2"
	ToolchainTestID  = "sha256:0be1f592782114f96b3422c3622cc57593c9db4e13dea376bd6e7065ccdb8acd"
	MaxExecutions    = 1
)

const CanonicalSource = "package valuewitness\nnamespace valuewitness\n\nentity Integer id \"gooo://value-witness/entity/integer\"\n\nactivity Increment(Integer) -> Integer computes \"int.add:1\"\n"

// SourceSpan is the immutable AST location of the activity declaration.
// Offsets make the span independently checkable against Source.Bytes; lines
// and columns keep the evidence human-auditable.
type SourceSpan struct {
	SourceID    string `json:"source_id"`
	StartByte   int    `json:"start_byte"`
	EndByte     int    `json:"end_byte"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column"`
}

type SourceSnapshot struct {
	Path   string `json:"path"`
	Bytes  string `json:"bytes"`
	Digest string `json:"digest"`
}

type ActivityIdentity struct {
	DeclarationKind     string     `json:"declaration_kind"`
	Name                string     `json:"name"`
	QualifiedName       string     `json:"qualified_name"`
	InputEntities       []string   `json:"input_entities"`
	OutputEntity        string     `json:"output_entity"`
	ValueProgram        string     `json:"value_program"`
	ValueProgramDigest  string     `json:"value_program_digest"`
	SemanticFingerprint string     `json:"semantic_fingerprint"`
	ASTSpan             SourceSpan `json:"ast_span"`
}

type ValueCase struct {
	ID             string `json:"id"`
	Input          int64  `json:"input"`
	ExpectedOutput int64  `json:"expected_output"`
}

type RegistryIdentity struct {
	Schema           string                         `json:"schema"`
	ID               string                         `json:"id"`
	Version          string                         `json:"version"`
	EvaluatorID      string                         `json:"evaluator_id"`
	EvaluatorVersion string                         `json:"evaluator_version"`
	Operations       []valueexecution.OperationSpec `json:"operations"`
	Digest           string                         `json:"digest"`
}

type ExecutionInput struct {
	Schema                  string           `json:"schema"`
	ContractID              string           `json:"contract_id"`
	CandidateStableID       string           `json:"candidate_stable_id"`
	CandidateDigest         string           `json:"candidate_digest"`
	SubjectSHA              string           `json:"subject_sha"`
	ObservationDigest       string           `json:"observation_digest"`
	OperationID             string           `json:"operation_id"`
	BoundedTarget           string           `json:"bounded_target"`
	Phase                   string           `json:"phase"`
	Source                  SourceSnapshot   `json:"source"`
	Activity                ActivityIdentity `json:"activity"`
	Corpus                  []ValueCase      `json:"corpus"`
	CorpusDigest            string           `json:"corpus_digest"`
	AllowedEffects          []string         `json:"allowed_effects"`
	EvaluatorRegistry       RegistryIdentity `json:"evaluator_registry"`
	ToolchainTestContractID string           `json:"toolchain_test_contract_identity"`
	OutputSchema            string           `json:"output_schema"`
	InputAuthority          string           `json:"input_authority"`
	OutputAuthority         string           `json:"output_authority"`
	MaxExecutions           int              `json:"max_executions"`
	RepositoryWritesAllowed bool             `json:"repository_writes_allowed"`
	Digest                  string           `json:"digest"`
}
