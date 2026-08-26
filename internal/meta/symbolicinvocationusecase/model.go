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
	ExpectedGeneratedCount       int    `json:"expected_generated_instances"`
	ExpectedGoldenMatches        int    `json:"expected_generated_golden_matches"`
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
	Schema               string             `json:"schema"`
	Decision             string             `json:"decision"`
	Resolution           string             `json:"resolution"`
	Reason               string             `json:"reason"`
	SubjectSHA           string             `json:"subject_sha"`
	Compiler             CompilerEvidence   `json:"compiler"`
	Source               SourceCoordinate   `json:"source"`
	Artifact             ArtifactEvidence   `json:"artifact"`
	Validation           ValidationEvidence `json:"validation"`
	DeterministicReplays int                `json:"deterministic_replays"`
	Resources            ResourceEvidence   `json:"resources"`
	Effects              Effects            `json:"effects"`
	NotClaimed           []string           `json:"not_claimed"`
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
	Kind              string `json:"kind"`
	ArtifactSchema    string `json:"artifact_schema"`
	Digest            string `json:"digest"`
	JSONSchemaDialect string `json:"json_schema_dialect"`
	JSONSchemaDigest  string `json:"json_schema_digest"`
}
