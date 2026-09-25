package selfimprovementcandidate

type sourceObservation struct {
	Schema              string            `json:"schema"`
	Metaprogram         string            `json:"metaprogram"`
	SubjectSHA          string            `json:"subject_sha"`
	SourceWorkflowRunID int64             `json:"source_workflow_run_id"`
	ContractID          string            `json:"contract_id"`
	Decision            string            `json:"decision"`
	Resolution          string            `json:"resolution"`
	Reason              string            `json:"reason"`
	Interpretation      string            `json:"interpretation"`
	Summary             sourceSummary     `json:"summary"`
	Authority           sourceAuthority   `json:"authority"`
	Artifacts           []sourceArtifact  `json:"artifacts"`
	InputDigest         string            `json:"input_digest"`
	Indicators          []sourceIndicator `json:"indicators"`
	Views               []sourceView      `json:"views"`
	Proofs              []sourceProof     `json:"proofs"`
	NotClaimed          []string          `json:"not_claimed"`
	Digest              string            `json:"digest"`
}

type sourceSummary struct {
	Coordinates         Coordinate `json:"coordinates"`
	SourceCoordinates   Coordinate `json:"source_coordinates"`
	Counterexamples     Coordinate `json:"counterexamples"`
	GoooDefinitionFiles int        `json:"gooo_definition_files"`
	GoDefinitionFiles   int        `json:"go_definition_files"`
	ResourceSamples     int        `json:"resource_samples"`
	MaxWallMS           int        `json:"max_wall_ms"`
	MaxRSSKiB           int        `json:"max_rss_kib"`
	BinaryBytes         int        `json:"binary_bytes"`
	CandidateCount      int        `json:"candidate_count"`
	Unknowns            int        `json:"unknowns"`
}

type sourceAuthority struct {
	RepositoryWrites            int  `json:"repository_writes"`
	MutationAuthorized          bool `json:"mutation_authorized"`
	ExecutionAuthorized         bool `json:"execution_authorized"`
	PromotionAuthorized         bool `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized bool `json:"automatic_adoption_authorized"`
}

type sourceArtifact struct {
	Kind           string `json:"kind"`
	Schema         string `json:"schema"`
	FileDigest     string `json:"file_digest"`
	SemanticDigest string `json:"semantic_digest"`
	Decision       string `json:"decision"`
}
