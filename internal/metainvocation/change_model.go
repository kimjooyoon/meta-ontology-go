package metainvocation

type ChangeSet struct {
	Schema string   `json:"schema"`
	CaseID string   `json:"case_id"`
	Files  []string `json:"files"`
}

type RuleEvidence struct {
	ID         string           `json:"id"`
	Operation  string           `json:"operation"`
	File       string           `json:"file"`
	SpecDigest string           `json:"spec_digest"`
	Source     SourceCoordinate `json:"source"`
}
