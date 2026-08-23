package languagesemanticbinding

type Indicator struct {
	MetricID     string `json:"metric_id"`
	Class        string `json:"class"`
	ProofChoice  string `json:"proof_choice"`
	Producer     string `json:"producer"`
	Consumer     string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Value        int    `json:"value"`
	Target       int    `json:"target"`
	Satisfied    bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}
