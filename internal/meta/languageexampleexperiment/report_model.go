package languageexampleexperiment

const ReportSchema = "gooo/language-example-experiment-report/v1"

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

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int64  `json:"value"`
	Target        int64  `json:"target"`
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
	Choice        string `json:"choice"`
	Claim         string `json:"claim"`
	MetaOperation string `json:"meta_operation"`
	Evidence      string `json:"evidence_digest"`
	Passed        bool   `json:"passed"`
}
