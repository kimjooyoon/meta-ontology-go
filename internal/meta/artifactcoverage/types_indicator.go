package artifactcoverage

type Indicator struct {
	MetricID      string         `json:"metric_id"`
	Class         IndicatorClass `json:"class"`
	Target        int            `json:"target"`
	Unit          string         `json:"unit"`
	Relation      Relation       `json:"relation"`
	ProofChoice   ProofChoice    `json:"proof_choice"`
	Producer      string         `json:"producer"`
	Consumer      string         `json:"consumer"`
	MetaOperation string         `json:"meta_operation"`
	Activity      string         `json:"activity"`
}
