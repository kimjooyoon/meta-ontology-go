package integrationprogress

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Value         int64  `json:"value"`
	Target        *int64 `json:"target,omitempty"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Status        string `json:"status"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
}

type Proof struct {
	Choice   string   `json:"choice"`
	Status   string   `json:"status"`
	Evidence []string `json:"evidence"`
}
