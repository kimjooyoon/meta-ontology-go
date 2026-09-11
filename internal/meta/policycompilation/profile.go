package policycompilation

const (
	PublicProfileID                    = "meta-policy-compilation-v3"
	PublicGenerationManifestSchema     = "gooo/meta-policy-compilation-generation-manifest/v1"
	PublicGenerationOutputRootClass    = "CALLER_OWNED_EXTERNAL"
	PublicGenerationConformanceUnknown = "UNKNOWN"
)

// PublicGenerationManifest binds the source-derived policy and the generated
// standalone judge without claiming that either has been executed. Execution
// and conformance are deliberately CI-owned concerns.
type PublicGenerationManifest struct {
	Schema               string   `json:"schema"`
	Profile              string   `json:"profile"`
	SourceFile           string   `json:"source_file"`
	SourceDigest         string   `json:"source_digest"`
	SemanticDigest       string   `json:"semantic_digest"`
	Package              string   `json:"package"`
	Namespace            string   `json:"namespace"`
	PolicyBytesDigest    string   `json:"policy_bytes_digest"`
	ArtifactBytesDigest  string   `json:"artifact_bytes_digest"`
	GeneratedJudgeDigest string   `json:"generated_judge_digest"`
	GeneratedFiles       []string `json:"generated_files"`
	OutputRootClass      string   `json:"output_root_class"`
	ExecutionObserved    bool     `json:"execution_observed"`
	CurrentConformance   string   `json:"current_conformance"`
	RepositoryWrites     int      `json:"repository_writes"`
	MutationAuthority    int      `json:"mutation_authority"`
	PromotionAuthority   int      `json:"promotion_authority"`
}
