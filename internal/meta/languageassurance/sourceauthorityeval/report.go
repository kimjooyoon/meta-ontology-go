package sourceauthorityeval

type Summary struct {
	AcceptedFacts int `json:"accepted_facts"`
	BackedFacts   int `json:"backed_facts"`
	UnknownFacts  int `json:"unknown_facts"`
	FailedFacts   int `json:"failed_facts"`
	ErrorFacts    int `json:"error_facts"`
	CoverageBPS   int `json:"coverage_bps"`
}

type FactReceipt struct {
	FactID       string `json:"fact_id"`
	SourceRef    string `json:"source_ref"`
	AuthorityRef string `json:"authority_ref"`
	Observation  string `json:"observation"`
	Resolution   string `json:"resolution"`
	Reason       string `json:"reason"`
}
