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
	identity    string
	kind        string
	lines       int
	movable     bool
	compactable bool
}

type planSubject struct {
	Logical      string `json:"logical"`
	Lines        int    `json:"lines"`
	RequiredSave int    `json:"required_savings"`
	MaxAtomLines int    `json:"max_atom_lines"`
	MovableAtoms int    `json:"movable_atoms"`
	DensityAtoms int    `json:"density_atoms"`
	Reason       string `json:"reason"`
	Consumer     string `json:"consumer"`
	Operation    string `json:"meta_operation"`
	Proof        string `json:"proof_choice"`
	Executable   bool   `json:"executable"`
}

type planCounterexample struct {
	Logical       string   `json:"logical"`
	BlockerID     string   `json:"blocker_id"`
	Decision      string   `json:"decision"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
	Diagnostics   []string `json:"diagnostics"`
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
	Schema          string              `json:"schema"`
	SourceSHA       string              `json:"source_sha"`
	Subjects        []planSubject       `json:"subjects"`
	Counterexamples []planCounterexample `json:"counterexamples"`
	Indicators      []planIndicator     `json:"indicators"`
}

type packageShape struct {
	BranchEntries int            `json:"branch_entries"`
	MaxEntries    int            `json:"max_entries"`
	Leaves        map[string]int `json:"leaves"`
}
