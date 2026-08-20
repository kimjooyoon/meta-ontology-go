package main

type splitPlan struct {
	Schema    string        `json:"schema"`
	SourceSHA string        `json:"source_sha"`
	Subjects  []planSubject `json:"subjects"`
}

type planSubject struct {
	Logical string `json:"logical"`
	Lines   int    `json:"lines"`
}

type densityReport struct {
	Schema    string           `json:"schema"`
	SourceSHA string           `json:"source_sha"`
	Subjects  []densitySubject `json:"subjects"`
}

type densitySubject struct {
	Logical string `json:"logical"`
	Status  string `json:"status"`
}

type extractionSubject struct {
	Logical   string   `json:"logical"`
	Before    int      `json:"before_lines"`
	After     int      `json:"after_lines"`
	Files     []string `json:"changed_files"`
	Consumer  string   `json:"consumer"`
	Operation string   `json:"meta_operation"`
	Proof     string   `json:"proof_choice"`
}

type extractionIndicator struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

type extractionReport struct {
	Schema     string                `json:"schema"`
	SourceSHA  string                `json:"source_sha"`
	Subjects   []extractionSubject   `json:"subjects"`
	Unhandled  []string              `json:"unhandled"`
	Indicators []extractionIndicator `json:"indicators"`
}
