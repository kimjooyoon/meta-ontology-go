package main

type admissionReport struct {
	Decision              string          `json:"decision"`
	CasesDeclared         int             `json:"cases_declared"`
	CasesPassed           int             `json:"cases_passed"`
	ElapsedNanoseconds    int64           `json:"elapsed_nanoseconds"`
	BaselineReceiptDigest string          `json:"baseline_receipt_digest"`
	BaselineCredit        bool            `json:"baseline_readiness_credit"`
	SetupError            string          `json:"setup_error,omitempty"`
	Cases                 []admissionCase `json:"cases"`
}

type admissionCase struct {
	ID                   string            `json:"id"`
	ExpectedReason       string            `json:"expected_reason"`
	SourceReason         string            `json:"source_reason"`
	FirstConsumerReason  string            `json:"first_consumer_reason"`
	ReplayConsumerReason string            `json:"replay_consumer_reason"`
	ReadinessCredit      bool              `json:"readiness_credit"`
	ResultProduced       bool              `json:"readiness_result_produced"`
	InputDigests         map[string]string `json:"input_digests"`
	HarnessError         string            `json:"harness_error,omitempty"`
	Passed               bool              `json:"passed"`
}
