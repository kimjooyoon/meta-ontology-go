package directorypartition

type Candidate struct {
	Subject                string `json:"subject"`
	IndicatorID            string `json:"indicator_id"`
	Limit                  int    `json:"limit"`
	EntryCount             int    `json:"entry_count"`
	BucketCount            int    `json:"bucket_count"`
	ProjectedDirectEntries int    `json:"projected_direct_entries"`
	Moves                  []Move `json:"moves"`
	Status                 string `json:"status"`
}

type Move struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Kind        string `json:"kind"`
	Bucket      int    `json:"bucket"`
}

type planCore struct {
	Summary    Summary     `json:"summary"`
	Indicators []Indicator `json:"indicators"`
	Candidates []Candidate `json:"candidates"`
	Proofs     []Proof     `json:"proofs"`
}
