package predecessorbinding

const (
	Schema         = "gooo/language-readiness-predecessor-binding/v1"
	RegistrySchema = "gooo/language-readiness-predecessor-coordinate-registry/v1"
	SourcePath     = "internal/meta/languagereadiness/artifact/baseline.go"
	Provider       = "FoundationBaseline"
	UseCase        = "ci-selects-accepted-readiness-predecessor"
	Total          = 8
)

const (
	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
)

const (
	ReasonExactlyCounted  = "PREDECESSOR_BINDINGS_EXACTLY_COUNTED"
	ReasonHeadUnknown     = "PREDECESSOR_BINDING_HEAD_UNKNOWN"
	ReasonEvidenceUnknown = "PREDECESSOR_BINDING_EVIDENCE_UNKNOWN"
	ReasonWriteEffect     = "PREDECESSOR_BINDING_WRITE_EFFECT"
)

type State string

const (
	StateStaticLiteral State = "STATIC_LITERAL"
	StateDynamicInput  State = "DYNAMIC_INPUT"
	StateUnknown       State = "UNKNOWN"
)

var coordinates = []Coordinate{
	{ID: "run-id", GoField: "RunID"},
	{ID: "artifact-name", GoField: "ArtifactName"},
	{ID: "head-sha", GoField: "HeadSHA"},
	{ID: "file-sha256", GoField: "FileSHA256"},
	{ID: "artifact-digest", GoField: "ArtifactDigest"},
	{ID: "snapshot-digest", GoField: "SnapshotDigest"},
	{ID: "completed", GoField: "Completed"},
	{ID: "basis-points", GoField: "BasisPoints"},
}

func Registry() []Coordinate {
	result := make([]Coordinate, len(coordinates))
	copy(result, coordinates)
	return result
}
