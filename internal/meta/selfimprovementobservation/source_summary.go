package selfimprovementobservation

type SourceValue struct {
	PrimaryArtifacts     int `json:"primary_artifacts"`
	ArtifactDigestChecks int `json:"artifact_digest_checks"`
	GoldenMatches        int `json:"golden_matches"`
	DeterministicReplays int `json:"deterministic_replays"`
}

type SourceCompiler struct {
	SourceFiles               int `json:"source_files"`
	GoooFiles                 int `json:"gooo_files"`
	GoFiles                   int `json:"go_files"`
	GoooDefinitionBasisPoints int `json:"gooo_definition_basis_points"`
	RegisteredEmitters        int `json:"registered_emitters"`
}

type SourceResources struct {
	Samples          int   `json:"samples"`
	ValidSamples     int   `json:"valid_samples"`
	MaxWallMS        int64 `json:"max_wall_ms"`
	MaxRSSKiB        int64 `json:"max_rss_kib"`
	BinaryBytes      int64 `json:"binary_bytes"`
	WallViolations   int   `json:"wall_violations"`
	RSSViolations    int   `json:"rss_violations"`
	BinaryViolations int   `json:"binary_violations"`
}

type SourceCounterexamples struct {
	UnknownEmitterRejections int `json:"unknown_emitter_rejections"`
}

type SourceEffects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}
