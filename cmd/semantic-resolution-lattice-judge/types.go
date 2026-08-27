package main

type observation struct {
	Required          int    `json:"required"`
	Observed          int    `json:"observed"`
	Reason            string `json:"reason"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
}

type unknownValue struct {
	Stage  string `json:"stage"`
	Step   int    `json:"step"`
	Reason string `json:"reason"`
}

type transition struct {
	FromResolution    string        `json:"from_resolution"`
	ToResolution      string        `json:"to_resolution,omitempty"`
	Decision          string        `json:"decision"`
	Reason            string        `json:"reason"`
	Unknown           *unknownValue `json:"unknown,omitempty"`
	RepositoryWrites  int           `json:"repository_writes"`
	MutationAuthority bool          `json:"mutation_authority"`
}

type latticeCase struct {
	ID          string      `json:"id"`
	Decision    string      `json:"decision"`
	Observation observation `json:"observation"`
	Transition  transition  `json:"transition"`
	ClaimID     string      `json:"claim_id"`
}

type claim struct {
	ID          string `json:"id"`
	State       string `json:"state"`
	BeforeState string `json:"before_state"`
	AfterState  string `json:"after_state"`
	Preserved   bool   `json:"preserved"`
}

type metric struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Proof         string `json:"proof"`
}

type receipt struct {
	Schema            string `json:"schema"`
	Source            string `json:"source"`
	SourceSHA256      string `json:"source_sha256"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
	CaseDenominator   int    `json:"case_denominator"`
	Counts            struct {
		CasesTotal int `json:"cases_total"`
		Pass       int `json:"pass"`
		FailClosed int `json:"fail_closed"`
		Unknown    int `json:"unknown"`
	} `json:"counts"`
	Cases   []latticeCase `json:"cases"`
	Claims  []claim       `json:"claims"`
	Metrics []metric      `json:"metrics"`
}
