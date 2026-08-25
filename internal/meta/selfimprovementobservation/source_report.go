package selfimprovementobservation

type SourceReport struct {
	Schema         string        `json:"schema"`
	Decision       string        `json:"decision"`
	Resolution     string        `json:"resolution"`
	Reason         string        `json:"reason"`
	Interpretation string        `json:"interpretation"`
	SubjectSHA     string        `json:"subject_sha"`
	ContractID     string        `json:"contract_id"`
	Summary        SourceSummary `json:"summary"`
	Indicators     []SourceIndicator `json:"indicators"`
	Views          []SourceView      `json:"views"`
	Proofs         []SourceProof     `json:"proofs"`
	NotClaimed     []string          `json:"not_claimed"`
	FactsDigest    string            `json:"facts_digest"`
	Digest         string            `json:"digest"`
}

type SourceSummary struct {
	Coordinates     CountSummary          `json:"coordinates"`
	Value           SourceValue           `json:"value"`
	Compiler        SourceCompiler        `json:"compiler"`
	Resources       SourceResources       `json:"resources"`
	Counterexamples SourceCounterexamples `json:"counterexamples"`
	Effects         SourceEffects         `json:"effects"`
	NotClaimed      int                   `json:"not_claimed"`
	Unknowns        int                   `json:"unknowns"`
}

type CountSummary struct {
	Satisfied  int `json:"satisfied"`
	Total      int `json:"total"`
	BasisPoints int `json:"basis_points"`
}
