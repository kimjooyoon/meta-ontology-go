package symbolicinvocationusecase

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type ProducerBinding struct {
	ReceiptSchema        string `json:"receipt_schema"`
	ArtifactSchema       string `json:"artifact_schema"`
	ArtifactDigest       string `json:"artifact_digest"`
	JSONSchemaDigest     string `json:"json_schema_digest"`
	Validator            string `json:"validator"`
	ValidatorDigest      string `json:"validator_digest"`
	CompilerBinaryBytes  int64  `json:"compiler_binary_bytes"`
	CompilerBinaryDigest string `json:"compiler_binary_digest"`
	RegisteredEmitters   int    `json:"registered_emitters"`
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
	GeneratedInstances   int                 `json:"generated_instances"`
	GoldenMatches        int                 `json:"generated_golden_matches"`
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
