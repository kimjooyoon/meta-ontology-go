package artifactresolutionexperiment

type Report struct {
	Schema         string      `json:"schema"`
	Decision       string      `json:"decision"`
	Resolution     string      `json:"resolution"`
	Reason         string      `json:"reason"`
	Interpretation string      `json:"interpretation"`
	SubjectSHA     string      `json:"subject_sha"`
	ContractID     string      `json:"contract_id"`
	Summary        Summary     `json:"summary"`
	Indicators     []Indicator `json:"indicators"`
	Views          []View      `json:"views"`
	Proofs         []Proof     `json:"proofs"`
	NotClaimed     []string    `json:"not_claimed"`
	FactsDigest    string      `json:"facts_digest"`
	Digest         string      `json:"digest"`
}
