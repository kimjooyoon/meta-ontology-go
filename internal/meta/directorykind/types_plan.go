package directorykind

type Candidate struct {
	Subject       string  `json:"subject"`
	IndicatorID   string  `json:"indicator_id"`
	ObservedKinds int     `json:"observed_kinds"`
	EntryCount    int     `json:"entry_count"`
	GroupCount    int     `json:"group_count"`
	Groups        []Group `json:"groups"`
	Moves         []Move  `json:"moves"`
	Status        string  `json:"status"`
}

type Group struct {
	Kind        string `json:"kind"`
	Destination string `json:"destination"`
	EntryCount  int    `json:"entry_count"`
}

type Move struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	EntryKind   string `json:"entry_kind"`
	Language    string `json:"language,omitempty"`
}

type planCore struct {
	Summary    Summary     `json:"summary"`
	Indicators []Indicator `json:"indicators"`
	Candidates []Candidate `json:"candidates"`
	Proofs     []Proof     `json:"proofs"`
}
