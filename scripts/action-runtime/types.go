package main

const (
	reportSchema    = "gooo/action-runtime-conformance/v1"
	metaprogramPath = "scripts/action-runtime"
)

type Rule struct {
	Action        string   `json:"action"`
	MinimumMajor  int      `json:"minimum_major"`
	Runtime       string   `json:"runtime"`
	AllowedInputs []string `json:"allowed_inputs,omitempty"`
	Evidence      string   `json:"evidence"`
}

type useSite struct {
	Action string
	Ref    string
	Line   int
	Inputs []string
}

type Observation struct {
	Action         string   `json:"action"`
	Reference      string   `json:"reference"`
	Line           int      `json:"line"`
	MinimumMajor   int      `json:"minimum_major"`
	Runtime        string   `json:"runtime"`
	RuntimeVerdict string   `json:"runtime_verdict"`
	Inputs         []string `json:"inputs,omitempty"`
	InvalidInputs  []string `json:"invalid_inputs,omitempty"`
	InputVerdict   string   `json:"input_verdict"`
}

type Indicator struct {
	ID       string `json:"id"`
	Route    string `json:"route"`
	Verdict  string `json:"verdict"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}

type Report struct {
	Schema             string        `json:"schema"`
	Metaprogram        string        `json:"metaprogram"`
	CommitSHA          string        `json:"commit_sha"`
	Workflow           string        `json:"workflow"`
	SourceSHA256       string        `json:"source_sha256"`
	PolicySHA256       string        `json:"policy_sha256"`
	Status             string        `json:"status"`
	Reason             string        `json:"reason"`
	ActionsTotal       int           `json:"actions_total"`
	ActionsKnown       int           `json:"actions_known"`
	ActionsCompliant   int           `json:"actions_compliant"`
	InvalidInputsTotal int           `json:"invalid_inputs_total"`
	Policy             []Rule        `json:"policy"`
	Observations       []Observation `json:"observations"`
	Indicators         []Indicator   `json:"indicators"`
}
