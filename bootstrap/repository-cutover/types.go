package main

type cutoverConfig struct {
	root        string
	physical    string
	authority   string
	expectedSHA string
	backup      string
	evidence    string
	apply       bool
}

type catalog struct {
	Schema    string         `json:"schema"`
	SourceSHA string         `json:"source_sha"`
	Entries   []catalogEntry `json:"entries"`
}

type catalogEntry struct {
	Backing string `json:"backing"`
}

type gitState struct {
	Head    string
	Dirty   int
	Tracked []string
}

type cutoverIndicator struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

type cutoverEvidence struct {
	Schema           string             `json:"schema"`
	SourceSHA        string             `json:"source_sha"`
	LogicalOriginSHA string             `json:"logical_origin_sha"`
	AuthoritySHA256  string             `json:"authority_sha256"`
	CandidateSHA256  string             `json:"candidate_sha256"`
	LogicalEntries   int                `json:"logical_entries"`
	PhysicalEntries  int                `json:"physical_entries"`
	PlannedPaths     int                `json:"planned_paths"`
	Applied          bool               `json:"applied"`
	Indicators       []cutoverIndicator `json:"indicators"`
}

func metric(id, operation, proof string, value int, blocking bool) cutoverIndicator {
	return cutoverIndicator{ID: id, Value: value, Limit: 0, Blocking: blocking,
		Consumer: "repository-cutover", Operation: operation, Proof: proof}
}
