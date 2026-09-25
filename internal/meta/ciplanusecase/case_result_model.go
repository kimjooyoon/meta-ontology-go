package ciplanusecase

import "github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"

type CaseResult struct {
	ID               string                        `json:"id"`
	ExpectedDecision string                        `json:"expected_decision"`
	ObservedDecision string                        `json:"observed_decision"`
	ProofChoice      string                        `json:"proof_choice"`
	Status           string                        `json:"status"`
	Unknowns         []metainvocation.UnknownCause `json:"unknowns"`
	ClaimStatuses    map[string]string             `json:"claim_statuses"`
	EvidenceDigest   string                        `json:"evidence_digest"`
}
