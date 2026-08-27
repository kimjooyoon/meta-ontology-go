package languageutility

type Indicator struct {
	ID          string `json:"id"`
	Class       string `json:"class"`
	ProofChoice string `json:"proof_choice"`
	Observed    int    `json:"observed"`
	Target      int    `json:"target"`
	Unit        string `json:"unit"`
	Status      string `json:"status"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Closed         int    `json:"closed"`
	Total          int    `json:"total"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Report struct {
	Schema            string           `json:"schema"`
	ContractID        string           `json:"contract_id"`
	SubjectSHA        string           `json:"subject_sha"`
	Decision          string           `json:"decision"`
	Resolution        string           `json:"resolution"`
	Reason            string           `json:"reason"`
	Summary           Summary          `json:"summary"`
	UseCases          []UseCaseSummary `json:"use_cases"`
	Cells             []CellResult     `json:"cells"`
	Indicators        []Indicator      `json:"indicators"`
	Proofs            []Proof          `json:"proofs"`
	NotClaimed        []string         `json:"not_claimed"`
	ContractDigest    string           `json:"contract_digest"`
	ObservationDigest string           `json:"observation_digest"`
	ProgramDigest     string           `json:"program_digest"`
	Digest            string           `json:"digest"`
}
