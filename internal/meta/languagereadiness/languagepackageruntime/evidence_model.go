package languagepackageruntime

type Indicator struct {
	MetricID      string     `json:"metric_id"`
	Class         string     `json:"class"`
	ProofChoice   string     `json:"proof_choice"`
	MetaOperation string     `json:"meta_operation"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	Value         int        `json:"value"`
	Target        int        `json:"target"`
	Resolution    Resolution `json:"resolution"`
	Satisfied     bool       `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Stage struct {
	Sequence      int    `json:"sequence"`
	Name          string `json:"name"`
	MetaOperation string `json:"meta_operation"`
	InputDigest   string `json:"input_digest"`
	OutputDigest  string `json:"output_digest"`
	Effectful     bool   `json:"effectful"`
}
