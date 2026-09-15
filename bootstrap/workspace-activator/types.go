package main

type activationConfig struct {
	root            string
	logical         string
	storage         string
	gitDir          string
	gitIndex        string
	expectedSHA     string
	materialization string
	evidence        string
}

type sourceIndicator struct {
	ID       string `json:"id"`
	Value    int    `json:"value"`
	Limit    int    `json:"limit"`
	Blocking bool   `json:"blocking"`
}

type sourceEvidence struct {
	Schema     string            `json:"schema"`
	CurrentSHA string            `json:"current_sha"`
	Entries    int               `json:"entries"`
	Restored   int               `json:"restored"`
	Indicators []sourceIndicator `json:"indicators"`
}

type activationIndicator struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

type activationEvidence struct {
	Schema           string                `json:"schema"`
	SourceSHA        string                `json:"source_sha"`
	PhysicalEntries  int                   `json:"physical_entries"`
	StoredEntries    int                   `json:"stored_entries"`
	LogicalEntries   int                   `json:"logical_entries"`
	ActivatedEntries int                   `json:"activated_entries"`
	Indicators       []activationIndicator `json:"indicators"`
}

func metric(id, operation, proof string, value int) activationIndicator {
	return activationIndicator{ID: id, Value: value, Limit: 0, Blocking: true,
		Consumer: "workspace-activator", Operation: operation, Proof: proof}
}
