package languagedelivery

type Contract struct {
	Schema        string       `json:"schema"`
	ContractID    string       `json:"contract_id"`
	Version       int          `json:"version"`
	Scope         string       `json:"scope"`
	AudienceOrder []Audience   `json:"audience_order"`
	Obligations   []Obligation `json:"obligations"`
	NotClaimed    []string     `json:"not_claimed"`
	References    []Reference  `json:"references"`
}

type Obligation struct {
	ID            string         `json:"id"`
	Audience      Audience       `json:"audience"`
	Class         IndicatorClass `json:"class"`
	Outcome       string         `json:"outcome"`
	Evidence      EvidenceRule   `json:"evidence"`
	MetaOperation string         `json:"meta_operation"`
	ProofChoice   ProofChoice    `json:"proof_choice"`
}

type EvidenceRule struct {
	Source  SourceName   `json:"source"`
	Kind    EvidenceKind `json:"kind"`
	ID      string       `json:"id,omitempty"`
	Counter string       `json:"counter,omitempty"`
	Target  int          `json:"target"`
}

type Reference struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Authority string `json:"authority"`
}
