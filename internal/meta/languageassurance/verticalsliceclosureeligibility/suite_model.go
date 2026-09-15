package verticalsliceclosureeligibility

type Definition struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedEffect     string `json:"expected_effect"`
	ExpectedReason     string `json:"expected_reason"`
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Passed     bool       `json:"passed"`
	Report     Report     `json:"report"`
}

type Suite struct {
	Schema              string       `json:"schema"`
	SubjectSHA          string       `json:"subject_sha"`
	DenominatorID       string       `json:"denominator_id"`
	DenominatorDigest   string       `json:"denominator_digest"`
	Decision            string       `json:"decision"`
	Resolution          string       `json:"resolution"`
	Cases               []CaseResult `json:"cases"`
	CasesTotal          int          `json:"cases_total"`
	CasesPassed         int          `json:"cases_passed"`
	EligibleExact       int          `json:"eligible_exact"`
	UnknownFailClosed   int          `json:"unknown_fail_closed"`
	InvariantFailClosed int          `json:"invariant_fail_closed"`
	CoverageBPS         int          `json:"coverage_bps"`
	RepositoryWrites    int          `json:"repository_writes"`
	PromotionApplied    int          `json:"promotion_applied"`
	SuiteDigest         string       `json:"suite_digest"`
}
