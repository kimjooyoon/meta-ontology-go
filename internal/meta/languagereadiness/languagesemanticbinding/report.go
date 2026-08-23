package languagesemanticbinding

type Report struct {
	Schema           string      `json:"schema"`
	Decision         string      `json:"decision"`
	ReasonCode       string      `json:"reason_code"`
	Resolution       string      `json:"resolution"`
	Source           Source      `json:"source"`
	Summary          Summary     `json:"summary"`
	Indicators       []Indicator `json:"indicators"`
	Proofs           []Proof     `json:"proofs"`
	RepositoryWrites int         `json:"repository_writes"`
	MutationAuthorized bool      `json:"mutation_authorized"`
	ReportDigest     string      `json:"report_digest"`
}

type Input struct {
	ExpectedHeadSHA string
	ReadinessPath   string
	ConceptPath     string
	SemanticPath    string
}
