package claimledger

import "encoding/json"

const (
	ContractSchema = "gooo/claim-ledger-contract/v1"
	ReportSchema   = "gooo/claim-ledger/v1"
)

type Coordinate struct {
	Stage string `json:"stage"`
	Step  string `json:"step"`
}

type EvidenceSpec struct {
	Paths    []string        `json:"paths"`
	Operator string          `json:"operator"`
	Expected json.RawMessage `json:"expected,omitempty"`
}

type ClaimSpec struct {
	ID             string        `json:"id"`
	Kind           string        `json:"kind"`
	Modality       string        `json:"modality"`
	Subject        string        `json:"subject"`
	Predicate      string        `json:"predicate"`
	Scope          string        `json:"scope"`
	ProofRoute     string        `json:"proof_route"`
	Coordinate     Coordinate    `json:"coordinate"`
	Evidence       *EvidenceSpec `json:"evidence,omitempty"`
	UnknownReason  string        `json:"unknown_reason,omitempty"`
	RefutedReason  string        `json:"refuted_reason,omitempty"`
	ExcludedReason string        `json:"excluded_reason,omitempty"`
}

type ProofRouteCounts struct {
	Foundation int `json:"foundation"`
	Coherence  int `json:"coherence"`
	Regression int `json:"regression"`
}

type ExpectedMetrics struct {
	FixedClaimTotal      int              `json:"fixed_claim_total"`
	InScopeClaimTotal    int              `json:"in_scope_claim_total"`
	DischargedTotal      int              `json:"discharged_total"`
	UnknownTotal         int              `json:"unknown_total"`
	RefutedTotal         int              `json:"refuted_total"`
	ExcludedTotal        int              `json:"excluded_total"`
	OpenClaimTotal       int              `json:"open_claim_total"`
	DischargeBasisPoints int              `json:"discharge_basis_points"`
	FalsePromotionCount  int              `json:"false_promotion_count"`
	ProofRoutes          ProofRouteCounts `json:"proof_routes"`
	ClaimSetDecision     string           `json:"claim_set_decision"`
	Resolution           string           `json:"resolution"`
}

type Contract struct {
	Schema   string          `json:"schema"`
	Metric   string          `json:"metric"`
	Expected ExpectedMetrics `json:"expected"`
	Claims   []ClaimSpec     `json:"claims"`
}
