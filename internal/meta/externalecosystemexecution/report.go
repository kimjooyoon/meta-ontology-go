package externalecosystemexecution

type Indicator struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	Reason      string `json:"reason"`
}

func indicator(c Criterion, status, reason string) Indicator {
	n := 0
	if status == "SATISFIED" {
		n = 1
	}
	return Indicator{c.ID, c.Kind, status, n, 1, reason}
}

type Proof struct {
	Mode   string `json:"mode"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func hasFailure(outcomes []Outcome) bool {
	for _, item := range outcomes {
		if item.Action == "fail" || item.Action == "build-fail" {
			return true
		}
	}
	return false
}

type Report struct {
	Schema                string           `json:"schema"`
	ContractVersion       string           `json:"contract_version"`
	DenominatorVersion    string           `json:"denominator_version"`
	DenominatorDigest     string           `json:"denominator_digest"`
	Decision              string           `json:"decision"`
	Resolution            string           `json:"resolution"`
	Reason                string           `json:"reason"`
	Completed             int              `json:"completed"`
	Total                 int              `json:"total"`
	BasisPoints           int              `json:"basis_points"`
	UnknownIndicators     int              `json:"unknown_indicators"`
	ExternalExecutions    int              `json:"external_executions"`
	RepositoryWrites      int              `json:"repository_writes"`
	OfficialMutationCount int              `json:"official_mutation_count"`
	PromotionCount        int              `json:"promotion_count"`
	Reference             ReferenceReceipt `json:"reference"`
	Indicators            []Indicator      `json:"indicators"`
	Proofs                []Proof          `json:"proofs"`
}

type SuiteCase struct {
	Name               string `json:"name"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ActualDecision     string `json:"actual_decision"`
	ActualResolution   string `json:"actual_resolution"`
	Passed             bool   `json:"passed"`
}

type SuiteReport struct {
	Schema            string      `json:"schema"`
	Passed            int         `json:"passed"`
	Total             int         `json:"total"`
	UnknownExpected   int         `json:"unknown_expected"`
	InvariantExpected int         `json:"invariant_expected"`
	Unresolved        int         `json:"unresolved"`
	Cases             []SuiteCase `json:"cases"`
}
