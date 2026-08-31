package assuranceeligibility

type parentReport struct {
	Schema                string `json:"schema"`
	ContractVersion       string `json:"contract_version"`
	Decision              string `json:"decision"`
	Resolution            string `json:"resolution"`
	Reason                string `json:"reason"`
	DenominatorVersion    string `json:"denominator_version"`
	DenominatorDigest     string `json:"denominator_digest"`
	Completed             int    `json:"completed"`
	Total                 int    `json:"total"`
	BasisPoints           int    `json:"basis_points"`
	ExternalExecutions    int    `json:"external_executions"`
	OfficialMutationCount int    `json:"official_mutation_count"`
	PromotionCount        int    `json:"promotion_count"`
	RepositoryWrites      int    `json:"repository_writes"`
	UnknownIndicators     int    `json:"unknown_indicators"`
}

type reference struct {
	Available    bool   `json:"available"`
	BindingExact bool   `json:"binding_exact"`
	URL          string `json:"url"`
	Commit       string `json:"commit"`
	Tree         string `json:"tree"`
}

type parentObservation struct {
	Schema                string    `json:"schema"`
	GoVersion             string    `json:"go_version"`
	Reference             reference `json:"reference"`
	OfficialMutationCount int       `json:"official_mutation_count"`
	PromotionCount        int       `json:"promotion_count"`
}

type parentSuite struct {
	Schema            string `json:"schema"`
	Passed            int    `json:"passed"`
	Total             int    `json:"total"`
	UnknownExpected   int    `json:"unknown_expected"`
	InvariantExpected int    `json:"invariant_expected"`
	Unresolved        int    `json:"unresolved"`
}
