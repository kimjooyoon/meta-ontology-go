package externalcapabilityexecution

type Suite struct {
	Schema              string      `json:"schema"`
	SubjectSHA          string      `json:"subject_sha"`
	Decision            string      `json:"decision"`
	Resolution          string      `json:"resolution"`
	Passed              int         `json:"passed"`
	Total               int         `json:"total"`
	CoverageBPS         int         `json:"coverage_bps"`
	ExactExpected       int         `json:"exact_expected"`
	UnknownExpected     int         `json:"unknown_expected"`
	InvariantExpected   int         `json:"invariant_expected"`
	RepositoryWrites    int         `json:"repository_writes"`
	ExternalExecutions  int         `json:"external_executions"`
	OfficialMutations   int         `json:"official_mutations"`
	PromotionCount      int         `json:"promotion_count"`
	Cases               []SuiteCase `json:"cases"`
	SuiteDigest         string      `json:"suite_digest"`
}
