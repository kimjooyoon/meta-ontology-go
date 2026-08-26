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
	FixedClaimTotal     int              `json:"fixed_claim_total"`
	InScopeClaimTotal   int              `json:"in_scope_claim_total"`
	DischargedTotal     int              `json:"discharged_total"`
	UnknownTotal        int              `json:"unknown_total"`
	RefutedTotal        int              `json:"refuted_total"`
	ExcludedTotal       int              `json:"excluded_total"`
	OpenClaimTotal      int              `json:"open_claim_total"`
	DischargeBasisPoints int             `json:"discharge_basis_points"`
	FalsePromotionCount int              `json:"false_promotion_count"`
	ProofRoutes         ProofRouteCounts `json:"proof_routes"`
	ClaimSetDecision    string           `json:"claim_set_decision"`
	Resolution          string           `json:"resolution"`
}

type Contract struct {
	Schema   string          `json:"schema"`
	Metric   string          `json:"metric"`
	Expected ExpectedMetrics `json:"expected"`
	Claims   []ClaimSpec     `json:"claims"`
}

type Evidence struct {
	ID                  string `json:"id"`
	ClaimID             string `json:"claim_id"`
	Status              string `json:"status"`
	SourcePath          string `json:"source_path"`
	SourceDigest        string `json:"source_digest"`
	ObservedValueDigest string `json:"observed_value_digest,omitempty"`
	ExpectedValueDigest string `json:"expected_value_digest,omitempty"`
}

type Claim struct {
	ID           string     `json:"id"`
	Kind         string     `json:"kind"`
	Modality     string     `json:"modality"`
	Subject      string     `json:"subject"`
	Predicate    string     `json:"predicate"`
	Scope        string     `json:"scope"`
	ProofRoute   string     `json:"proof_route"`
	Coordinate   Coordinate `json:"coordinate"`
	Status       string     `json:"status"`
	Truth        string     `json:"truth"`
	Reason       string     `json:"reason"`
	EvidenceRefs []string   `json:"evidence_refs"`
}

type Event struct {
	Sequence   int        `json:"sequence"`
	Type       string     `json:"type"`
	ClaimID    string     `json:"claim_id"`
	EvidenceID string     `json:"evidence_id,omitempty"`
	Status     string     `json:"status"`
	Coordinate Coordinate `json:"coordinate"`
	Reason     string     `json:"reason"`
}

type Metrics struct {
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
}

type Verdict struct {
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
}

type Conformance struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type Report struct {
	Schema            string      `json:"schema"`
	Subject           string      `json:"subject"`
	Metric            string      `json:"metric"`
	ContractDigest    string      `json:"contract_digest"`
	ObservationDigest string      `json:"observation_digest"`
	ClaimSet          Verdict     `json:"claim_set"`
	Conformance       Conformance `json:"conformance"`
	Metrics           Metrics     `json:"metrics"`
	OpenClaimIDs      []string    `json:"open_claim_ids"`
	Claims            []Claim     `json:"claims"`
	Evidence          []Evidence  `json:"evidence"`
	Events            []Event     `json:"events"`
}
