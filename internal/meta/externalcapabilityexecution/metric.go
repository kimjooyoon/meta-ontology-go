package externalcapabilityexecution

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Resolution    string `json:"resolution"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Status        string `json:"status"`
}

type Proof struct {
	Mode       string `json:"mode"`
	Status     string `json:"status"`
	Completed  int    `json:"completed"`
	Total      int    `json:"total"`
	Resolution string `json:"resolution"`
}
