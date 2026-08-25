package symbolicinvocationusecase

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
	Schema            string  `json:"schema"`
	Decision          string  `json:"decision"`
	Resolution        string  `json:"resolution"`
	Reason            string  `json:"reason"`
	SubjectSHA        string  `json:"subject_sha"`
	ArtifactDigest    string  `json:"artifact_digest"`
	JSONSchemaDigest  string  `json:"json_schema_digest"`
	ToolDigest        string  `json:"tool_digest"`
	AcceptedInstances int     `json:"accepted_instances"`
	RejectedInstances int     `json:"rejected_instances"`
	Effects           Effects `json:"effects"`
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
