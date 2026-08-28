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
	Logical      string   `json:"logical"`
	Before       int      `json:"before_lines"`
	After        int      `json:"after_lines"`
	Files        []string `json:"changed_files"`
	CreatedFiles []string `json:"created_files,omitempty"`
	Consumer     string   `json:"consumer"`
	Operation    string   `json:"meta_operation"`
	Proof        string   `json:"proof_choice"`
}

type extractionFailureRecord struct {
	Logical string `json:"logical"`
	Decision string `json:"decision"`
	Stage string `json:"stage"`
	Step string `json:"step"`
	Reason string `json:"reason"`
	UnknownClass string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
	BlockedBy []string `json:"blocked_by"`
	Diagnostics []string `json:"diagnostics,omitempty"`
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
	Failures   []extractionFailureRecord `json:"failures,omitempty"`
	Indicators []extractionIndicator `json:"indicators"`
}
