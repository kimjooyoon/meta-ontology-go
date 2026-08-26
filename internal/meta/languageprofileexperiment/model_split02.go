package languageprofileexperiment

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema                  string      `json:"schema"`
	Decision                string      `json:"decision"`
	Resolution              string      `json:"resolution"`
	Reason                  string      `json:"reason"`
	Interpretation          string      `json:"interpretation"`
	SubjectSHA              string      `json:"subject_sha"`
	ContractID              string      `json:"contract_id"`
	ResourceObservationMode string      `json:"resource_observation_mode"`
	Summary                 Summary     `json:"summary"`
	Indicators              []Indicator `json:"indicators"`
	Views                   []View      `json:"views"`
	Proofs                  []Proof     `json:"proofs"`
	NotClaimed              []string    `json:"not_claimed"`
	RepositoryWrites        int         `json:"repository_writes"`
	MutationAuthority       bool        `json:"mutation_authority"`
	FactsDigest             string      `json:"facts_digest"`
	Digest                  string      `json:"digest"`
}
