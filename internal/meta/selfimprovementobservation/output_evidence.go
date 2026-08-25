package selfimprovementobservation

type Authority struct {
	RepositoryWrites           int  `json:"repository_writes"`
	MutationAuthorized         bool `json:"mutation_authorized"`
	ExecutionAuthorized        bool `json:"execution_authorized"`
	PromotionAuthorized        bool `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized bool `json:"automatic_adoption_authorized"`
}

type ArtifactRef struct {
	Kind           string `json:"kind"`
	Schema         string `json:"schema"`
	FileDigest     string `json:"file_digest"`
	SemanticDigest string `json:"semantic_digest"`
	Decision       string `json:"decision"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
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

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}
