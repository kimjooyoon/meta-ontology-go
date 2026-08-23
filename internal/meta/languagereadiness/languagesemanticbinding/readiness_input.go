package languagesemanticbinding

type readinessArtifact struct {
	Schema          string              `json:"schema"`
	HeadSHA         string              `json:"head_sha"`
	ArtifactDigest  string              `json:"artifact_digest"`
	Snapshot        readinessSnapshot   `json:"snapshot"`
	TransitionInput readinessTransition `json:"transition_input"`
}

type readinessSnapshot struct {
	Schema               string                `json:"schema"`
	ContractSchema       string                `json:"contract_schema"`
	Decision             string                `json:"decision"`
	RegistryDigest       string                `json:"registry_digest"`
	SourceArtifactDigest string                `json:"source_artifact_digest"`
	Summary              readinessSummary      `json:"summary"`
	Obligations          []readinessObligation `json:"obligations"`
	RepositoryWrites     int                   `json:"repository_writes"`
}

type readinessSummary struct {
	Completed        int `json:"completed"`
	Total            int `json:"total"`
	NotSatisfied     int `json:"not_satisfied"`
	Unresolved       int `json:"unresolved"`
	ReadinessBPS     int `json:"readiness_bps"`
	RatioNumerator   int `json:"ratio_numerator"`
	RatioDenominator int `json:"ratio_denominator"`
}

type readinessObligation struct {
	ID             string `json:"id"`
	Area           string `json:"area"`
	ProofChoice    string `json:"proof_choice"`
	ConceptID      string `json:"concept_id"`
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
}

type readinessTransition struct {
	ContractSchema string               `json:"contract_schema"`
	RegistryDigest string               `json:"registry_digest"`
	Completed      int                  `json:"completed"`
	Total          int                  `json:"total"`
	BasisPoints    int                  `json:"basis_points"`
	Evidence       []transitionEvidence `json:"evidence"`
}

type transitionEvidence struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
