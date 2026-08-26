package externalconformanceactivation

type eligibilityIndicator struct {
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Resolution    string `json:"resolution"`
	Satisfied     bool   `json:"satisfied"`
}
