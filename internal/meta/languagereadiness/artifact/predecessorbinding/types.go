package predecessorbinding

type Coordinate struct {
	ID      string `json:"id"`
	GoField string `json:"go_field"`
}

type Observation struct {
	ID         string `json:"id"`
	GoField    string `json:"go_field"`
	SourcePath string `json:"source_path"`
	Provider   string `json:"provider"`
	State      State  `json:"state"`
}

type Evidence struct {
	Observation
	Reason string `json:"reason"`
}

type Summary struct {
	StaticLiteral int `json:"static_literal"`
	DynamicInput  int `json:"dynamic_input"`
	Unknown       int `json:"unknown"`
	Total         int `json:"total"`
	DynamicBPS    int `json:"dynamic_bps"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Unit          string `json:"unit"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	ID     string `json:"id"`
	Choice string `json:"choice"`
	Passed bool   `json:"passed"`
}

type Report struct {
	Schema           string      `json:"schema"`
	RegistrySchema   string      `json:"registry_schema"`
	RegistryDigest   string      `json:"registry_digest"`
	HeadSHA          string      `json:"head_sha"`
	UseCase          string      `json:"use_case"`
	Decision         string      `json:"decision"`
	Reason           string      `json:"reason"`
	Summary          Summary     `json:"summary"`
	Evidence         []Evidence  `json:"evidence"`
	Indicators       []Indicator `json:"indicators"`
	Proofs           []Proof     `json:"proofs"`
	RepositoryWrites int         `json:"repository_writes"`
	ReportDigest     string      `json:"report_digest"`
}
