package ciplanusecase

const (
	ContractSchema = "gooo/ci-plan-usecase-contract/v1"
	ReportSchema   = "gooo/ci-plan-usecase-scorecard/v1"
)

type CaseSpec struct {
	ID               string `json:"id"`
	ExpectedDecision string `json:"expected_decision"`
	ProofChoice      string `json:"proof_choice"`
}

type Limits struct {
	MaxWallMS       int64 `json:"max_wall_ms"`
	MaxPeakRSSKiB   int64 `json:"max_peak_rss_kib"`
	MaxReceiptBytes int64 `json:"max_receipt_bytes"`
}

type Contract struct {
	Schema      string     `json:"schema"`
	Denominator int        `json:"denominator"`
	Cases       []CaseSpec `json:"cases"`
	Limits      Limits     `json:"limits"`
	NotClaimed  []string   `json:"not_claimed"`
}
