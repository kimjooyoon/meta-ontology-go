package operationconformance

type IndicatorObservation struct {
	ID                string   `json:"id"`
	Role              string   `json:"role"`
	Route             string   `json:"route"`
	RuleID            string   `json:"rule_id"`
	Decision          Decision `json:"decision"`
	Resolution        string   `json:"resolution"`
	Value             int      `json:"value"`
	Target            int      `json:"target"`
	ObservationDigest string   `json:"observation_digest"`
}

type Summary struct {
	PassCount                      int `json:"pass_count"`
	FailCount                      int `json:"fail_count"`
	UnknownCount                   int `json:"unknown_count"`
	Total                          int `json:"total"`
	EvaluatedIndicatorCount        int `json:"evaluated_indicator_count"`
	RuntimeObservedIndicatorCount  int `json:"runtime_observed_indicator_count"`
	IndependentImplementationCount int `json:"independent_implementation_count"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	Passed         bool   `json:"passed"`
	EvidenceDigest string `json:"evidence_digest"`
}

type IndicatorCounterexample struct {
	IndicatorID   string   `json:"indicator_id"`
	RuleID        string   `json:"rule_id"`
	Observed      int      `json:"observed"`
	Expected      int      `json:"expected"`
	Decision      Decision `json:"decision"`
	EvidenceDigest string  `json:"evidence_digest"`
}

type Report struct {
	Schema                       string                    `json:"schema"`
	ContractID                   string                    `json:"contract_id"`
	OperationID                  string                    `json:"operation_id"`
	Decision                     Decision                  `json:"decision"`
	Reason                       string                    `json:"reason"`
	Resolution                   string                    `json:"resolution"`
	AssuranceGrade               string                    `json:"assurance_grade"`
	MetaOperation                string                    `json:"meta_operation"`
	Contract                     ContractReceipt           `json:"contract"`
	Evidence                     SplitGoEvidence           `json:"evidence"`
	EvidenceDigest               string                    `json:"evidence_digest"`
	Summary                      Summary                   `json:"summary"`
	Indicators                   []IndicatorObservation    `json:"indicators"`
	Counterexamples              []IndicatorCounterexample `json:"counterexamples"`
	Proofs                       []Proof                   `json:"proofs"`
	RepositoryWrites             int                       `json:"repository_writes"`
	RepositoryMutationAuthorized bool                      `json:"repository_mutation_authorized"`
	ReportDigest                 string                    `json:"report_digest"`
}
