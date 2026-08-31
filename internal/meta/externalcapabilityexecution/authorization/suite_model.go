package authorization

type SuiteCase struct {
	CaseID             string `json:"case_id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ActualDecision     string `json:"actual_decision"`
	ActualResolution   string `json:"actual_resolution"`
	Passed             bool   `json:"passed"`
}

type Suite struct {
	Schema          string      `json:"schema"`
	SubjectSHA      string      `json:"subject_sha"`
	Decision        string      `json:"decision"`
	Resolution      string      `json:"resolution"`
	Passed          int         `json:"passed"`
	Total           int         `json:"total"`
	CoverageBPS     int         `json:"coverage_bps"`
	AuthorizedCases int         `json:"authorized_cases"`
	UnknownCases    int         `json:"unknown_cases"`
	DeniedCases     int         `json:"denied_cases"`
	Cases           []SuiteCase `json:"cases"`
	SuiteDigest     string      `json:"suite_digest"`
}
