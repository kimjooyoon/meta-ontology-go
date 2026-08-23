package operationconformance

type contractDocument struct {
	Schema, ContractID, OperationID string
	Denominator                     struct {
		Version          string         `json:"version"`
		IndicatorCount   int            `json:"indicator_count"`
		OracleCaseCount  int            `json:"oracle_case_count"`
		ExpectedOutcomes map[string]int `json:"expected_outcomes"`
	} `json:"denominator"`
	DecisionPolicy struct {
		Unknown string `json:"unknown"`
		Pass    struct {
			PassCount, FailCount, UnknownCount int
			Decision                           string `json:"decision"`
		} `json:"pass"`
		Otherwise string `json:"otherwise"`
	} `json:"decision_policy"`
	Indicators  []IndicatorDefinition `json:"indicators"`
	OracleCases []struct {
		ID, IndicatorID, Expected string
		Observation               map[string]any `json:"observation"`
	} `json:"oracle_cases"`
}

type ContractReceipt struct {
	Decision      Decision `json:"decision"`
	ContractID    string   `json:"contract_id"`
	Version       string   `json:"denominator_version"`
	ContractDigest string  `json:"contract_digest"`
	Matched       int      `json:"matched"`
	Total         int      `json:"total"`
	PassCases     int      `json:"pass_cases"`
	FailCases     int      `json:"fail_cases"`
	UnknownCases  int      `json:"unknown_cases"`
}
