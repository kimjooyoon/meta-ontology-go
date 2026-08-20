package main

type projectionEvidence struct {
	Schema    string         `json:"schema"`
	SourceSHA string         `json:"source_sha"`
	Subjects  []inputSubject `json:"subjects"`
}

type inputSubject struct {
	Indicator string `json:"indicator"`
	Logical   string `json:"logical"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
}

type declarationAtom struct {
	kind    string
	lines   int
	movable bool
}

type planSubject struct {
	Logical      string `json:"logical"`
	Lines        int    `json:"lines"`
	MaxAtomLines int    `json:"max_atom_lines"`
	MovableAtoms int    `json:"movable_atoms"`
	Reason       string `json:"reason"`
	Consumer     string `json:"consumer"`
	Operation    string `json:"meta_operation"`
	Proof        string `json:"proof_choice"`
}

type planIndicator struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

type planReport struct {
	Schema     string          `json:"schema"`
	SourceSHA  string          `json:"source_sha"`
	Subjects   []planSubject   `json:"subjects"`
	Indicators []planIndicator `json:"indicators"`
}
