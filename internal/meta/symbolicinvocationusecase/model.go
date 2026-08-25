package symbolicinvocationusecase

type Contract struct {
	Schema                       string `json:"schema"`
	Version                      int    `json:"version"`
	MetricID                     string `json:"metric_id"`
	ExpectedGoVersion            string `json:"expected_go_version"`
	ExpectedRegisteredEmitters   int    `json:"expected_registered_emitters"`
	ExpectedGoooFiles            int    `json:"expected_gooo_files"`
	ExpectedGoFiles              int    `json:"expected_go_files"`
	ExpectedGoooLines            int    `json:"expected_gooo_lines"`
	ExpectedFiles                int    `json:"expected_files"`
	ExpectedDirectories          int    `json:"expected_directories"`
	ExpectedAcceptedInstances    int    `json:"expected_accepted_instances"`
	ExpectedRejectedInstances    int    `json:"expected_rejected_instances"`
	ExpectedDeterministicReplays int    `json:"expected_deterministic_replays"`
	ExpectedResourceSamples      int    `json:"expected_resource_samples"`
	ExpectedRepositoryWrites     int    `json:"expected_repository_writes"`
	ExpectedMutationAuthorities  int    `json:"expected_mutation_authorities"`
	ExpectedNonClaims            int    `json:"expected_non_claims"`
	ExpectedValidator            string `json:"expected_validator"`
}

type Input struct {
	SubjectSHA       string           `json:"subject_sha"`
	Contract         Contract         `json:"contract"`
	ProducerReceipt  ProducerReceipt  `json:"producer_receipt"`
	ProducerArtifact ProducerArtifact `json:"producer_artifact"`
	Observation      Observation      `json:"observation"`
}

type ProducerReceipt struct {
	Schema               string            `json:"schema"`
	Decision             string            `json:"decision"`
	Resolution           string            `json:"resolution"`
	Reason               string            `json:"reason"`
	SubjectSHA           string            `json:"subject_sha"`
	Compiler             CompilerEvidence  `json:"compiler"`
	Source               SourceCoordinate  `json:"source"`
	Artifact             ArtifactEvidence  `json:"artifact"`
	Validation           ValidationEvidence `json:"validation"`
	DeterministicReplays int               `json:"deterministic_replays"`
	Resources            ResourceEvidence  `json:"resources"`
	Effects              Effects           `json:"effects"`
	NotClaimed           []string          `json:"not_claimed"`
}

type CompilerEvidence struct {
	GoVersion          string `json:"go_version"`
	BinaryDigest       string `json:"binary_digest"`
	BinaryBytes        int64  `json:"binary_bytes"`
	RegisteredEmitters int    `json:"registered_emitters"`
}

type SourceCoordinate struct {
	GoooFiles   int `json:"gooo_files"`
	GoFiles     int `json:"go_files"`
	GoooLines   int `json:"gooo_lines"`
	Files       int `json:"files"`
	Directories int `json:"directories"`
}

type ArtifactEvidence struct {
	Kind             string `json:"kind"`
	ArtifactSchema   string `json:"artifact_schema"`
	Digest           string `json:"digest"`
	JSONSchemaDialect string `json:"json_schema_dialect"`
	JSONSchemaDigest string `json:"json_schema_digest"`
}

type ValidationEvidence struct {
	Tool              string `json:"tool"`
	ToolDigest        string `json:"tool_digest"`
	AcceptedInstances int    `json:"accepted_instances"`
	RejectedInstances int    `json:"rejected_instances"`
}

type ResourceEvidence struct {
	Samples     []ResourceSample `json:"samples"`
	SampleCount int              `json:"sample_count"`
	MaxWallMS   int              `json:"max_wall_ms"`
	MaxRSSKiB   int              `json:"max_rss_kib"`
}

type ResourceSample struct {
	Sequence int `json:"sequence"`
	WallMS   int `json:"wall_ms"`
	RSSKiB   int `json:"rss_kib"`
}

type ProducerArtifact struct {
	Schema     string             `json:"schema"`
	Decision   string             `json:"decision"`
	Resolution string             `json:"resolution"`
	Reason     string             `json:"reason"`
	Kind       string             `json:"kind"`
	Extensions ArtifactExtensions `json:"extensions"`
	Effects    Effects            `json:"effects"`
	Digest     string             `json:"digest"`
}

type ArtifactExtensions struct {
	RegisteredEmitters int      `json:"registered_emitters"`
	Kinds              []string `json:"kinds"`
}

type Observation struct {
	Schema              string  `json:"schema"`
	Decision            string  `json:"decision"`
	Resolution          string  `json:"resolution"`
	Reason              string  `json:"reason"`
	SubjectSHA          string  `json:"subject_sha"`
	ArtifactDigest      string  `json:"artifact_digest"`
	JSONSchemaDigest    string  `json:"json_schema_digest"`
	ToolDigest          string  `json:"tool_digest"`
	AcceptedInstances   int     `json:"accepted_instances"`
	RejectedInstances   int     `json:"rejected_instances"`
	Effects             Effects `json:"effects"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Counter struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type ProducerBinding struct {
	ReceiptSchema       string `json:"receipt_schema"`
	ArtifactSchema      string `json:"artifact_schema"`
	ArtifactDigest      string `json:"artifact_digest"`
	JSONSchemaDigest    string `json:"json_schema_digest"`
	Validator           string `json:"validator"`
	ValidatorDigest     string `json:"validator_digest"`
	CompilerBinaryBytes int64  `json:"compiler_binary_bytes"`
	CompilerBinaryDigest string `json:"compiler_binary_digest"`
	RegisteredEmitters  int    `json:"registered_emitters"`
}

type ResourceObservation struct {
	Mode                       string `json:"mode"`
	MeasurementReplayAuthority bool   `json:"measurement_replay_authority"`
	Samples                    int    `json:"samples"`
	MaxWallMS                  int    `json:"max_wall_ms"`
	MaxRSSKiB                  int    `json:"max_rss_kib"`
}

type Summary struct {
	Coordinates          Counter             `json:"coordinates"`
	UserDecisions        int                 `json:"user_decisions"`
	AcceptedInstances    int                 `json:"accepted_instances"`
	RejectedInstances    int                 `json:"rejected_instances"`
	DeterministicReplays int                 `json:"deterministic_replays"`
	Unknowns             int                 `json:"unknowns"`
	Source               SourceCoordinate    `json:"source"`
	Producer             ProducerBinding     `json:"producer"`
	Resources            ResourceObservation `json:"resources"`
	Effects              Effects             `json:"effects"`
}

type Report struct {
	Schema             string      `json:"schema"`
	SubjectSHA         string      `json:"subject_sha"`
	MetricID           string      `json:"metric_id"`
	Decision           string      `json:"decision"`
	Resolution         string      `json:"resolution"`
	Reason             string      `json:"reason"`
	Summary            Summary     `json:"summary"`
	Indicators         []Indicator `json:"indicators"`
	Views              []View      `json:"views"`
	PromotionCreditBPS int         `json:"promotion_credit_bps"`
	RepositoryWrites   int         `json:"repository_writes"`
	MutationAuthority  bool        `json:"mutation_authority"`
	NotClaimed         []string    `json:"not_claimed"`
	Digest             string      `json:"digest"`
}

type facts struct {
	UserDecisions        int
	AcceptedInstances    int
	RejectedInstances    int
	DeterministicReplays int
	Unknowns             int
	Source               SourceCoordinate
	Producer             ProducerBinding
	Resources            ResourceObservation
	Effects              Effects
}
