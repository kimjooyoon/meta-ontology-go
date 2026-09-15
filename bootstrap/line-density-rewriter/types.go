package main

type splitPlan struct {
	Schema    string         `json:"schema"`
	SourceSHA string         `json:"source_sha"`
	Subjects  []splitSubject `json:"subjects"`
}

type splitSubject struct {
	Logical      string `json:"logical"`
	Lines        int    `json:"lines"`
	RequiredSave int    `json:"required_savings"`
	Reason       string `json:"reason"`
	Operation    string `json:"meta_operation"`
	Executable   bool   `json:"executable"`
}

type rewriteSubject struct {
	Logical    string `json:"logical"`
	Before     int    `json:"before_lines"`
	After      int    `json:"after_lines"`
	Operations int    `json:"operations"`
	Status     string `json:"status"`
	Consumer   string `json:"consumer"`
	Operation  string `json:"meta_operation"`
	Proof      string `json:"proof_choice"`
}

type rewriteIndicator struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

type rewriteReport struct {
	Schema     string             `json:"schema"`
	SourceSHA  string             `json:"source_sha"`
	Subjects   []rewriteSubject   `json:"subjects"`
	Indicators []rewriteIndicator `json:"indicators"`
}

type sourceSpan struct {
	start       int
	end         int
	replacement string
}
