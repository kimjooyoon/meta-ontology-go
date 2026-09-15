package languagedelivery

type Indicator struct {
	MetricID      string         `json:"metric_id"`
	Class         IndicatorClass `json:"class"`
	ProofChoice   ProofChoice    `json:"proof_choice"`
	MetaOperation string         `json:"meta_operation"`
	Value         int            `json:"value"`
	Target        int            `json:"target"`
	Satisfied     bool           `json:"satisfied"`
}

type Proof struct {
	Choice         ProofChoice `json:"choice"`
	Claim          string      `json:"claim"`
	MetaOperation  string      `json:"meta_operation"`
	EvidenceDigest string      `json:"evidence_digest"`
	Passed         bool        `json:"passed"`
}

type Report struct {
	Schema         string              `json:"schema"`
	Decision       string              `json:"decision"`
	Resolution     string              `json:"resolution"`
	Reason         string              `json:"reason"`
	SubjectSHA     string              `json:"subject_sha"`
	ContractID     string              `json:"contract_id"`
	ContractDigest string              `json:"contract_digest"`
	ManifestDigest string              `json:"manifest_digest"`
	Summary        Summary             `json:"summary"`
	Sources        []SourceObservation `json:"sources"`
	Obligations    []ObligationResult  `json:"obligations"`
	Views          []AudienceView      `json:"views"`
	Indicators     []Indicator         `json:"indicators"`
	Proofs         []Proof             `json:"proofs"`
	NotClaimed     []string            `json:"not_claimed"`
	FactsDigest    string              `json:"facts_digest"`
	Digest         string              `json:"digest"`
}
