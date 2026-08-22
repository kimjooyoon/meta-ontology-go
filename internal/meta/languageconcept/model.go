package languageconcept

import "io/fs"

const ReportSchema = "gooo/language-concept-catalog-report/v1"

type UseCase struct {
	ID              string `json:"id"`
	Trigger         string `json:"trigger"`
	ExpectedOutcome string `json:"expected_outcome"`
}

type Concept struct {
	ID             string    `json:"id"`
	Problem        string    `json:"problem"`
	PositiveEffect string    `json:"positive_effect"`
	MetaOperation  string    `json:"meta_operation"`
	Rarity         string    `json:"rarity"`
	Stage          string    `json:"stage"`
	NoveltyClaim   bool      `json:"novelty_claim"`
	CodeBindings   []string  `json:"code_bindings"`
	MetricBindings []string  `json:"metric_bindings"`
	UseCases       []UseCase `json:"use_cases"`
}

type Summary struct {
	Concepts              int `json:"concepts"`
	CodeBound             int `json:"code_bound"`
	UseCaseBound          int `json:"use_case_bound"`
	MetricBound           int `json:"metric_bound"`
	Operating             int `json:"operating"`
	Conformed             int `json:"conformed"`
	Unbound               int `json:"unbound"`
	UnverifiedNovelty     int `json:"unverified_novelty_claims"`
	RepositoryWrites      int `json:"repository_writes"`
}

type Report struct {
	Schema          string      `json:"schema"`
	Decision        string      `json:"decision"`
	Reason          string      `json:"reason"`
	Concepts        []Concept   `json:"concepts"`
	Summary         Summary     `json:"summary"`
	MissingBindings []string    `json:"missing_bindings"`
	Indicators      []Indicator `json:"indicators"`
	Proofs          []Proof     `json:"proofs"`
	ReportDigest    string      `json:"report_digest"`
}

func Evaluate(repository fs.FS, concepts []Concept) Report {
	return evaluate(repository, concepts)
}
