package workgraph

type Indicator struct {
	ID          string `json:"id"`
	Class       string `json:"class"`
	Value       int64  `json:"value"`
	Total       int64  `json:"total"`
	Unit        string `json:"unit"`
	Relation    string `json:"relation"`
	Target      int64  `json:"target"`
	Activity    string `json:"activity"`
	ProofChoice string `json:"proof_choice"`
	State       string `json:"state"`
}

type Report struct {
	Schema          string         `json:"schema"`
	HeadSHA         string         `json:"head_sha"`
	Project         string         `json:"project"`
	Decision        string         `json:"decision"`
	Resolution      string         `json:"resolution"`
	Reason          string         `json:"reason"`
	NextOperation   string         `json:"next_operation"`
	MutationAllowed bool           `json:"mutation_allowed"`
	ContractDigest  string         `json:"contract_digest"`
	SourceDigest    string         `json:"source_digest"`
	GeneratedDigest string         `json:"generated_digest,omitempty"`
	ReplayDigest    string         `json:"replay_digest,omitempty"`
	Cells           []Cell         `json:"cells"`
	Claim           ClaimLifecycle `json:"claim"`
	Resource        ResourceSample `json:"resource"`
	Summary         Summary        `json:"summary"`
	Indicators      []Indicator    `json:"indicators"`
}
