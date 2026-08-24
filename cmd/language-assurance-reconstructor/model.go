package main

import "encoding/json"

type authorityRoute struct {
	RuleID     string `json:"rule_id"`
	AuthoredBy string `json:"authored_by"`
	PromotedBy string `json:"promoted_by"`
}

type roleBinding struct {
	Principal string   `json:"principal"`
	Roles     []string `json:"roles"`
}
type decisionTransition struct {
	ID     string `json:"id"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type snapshotBinding struct {
	EvidenceID string `json:"evidence_id"`
	SubjectSHA string `json:"subject_sha"`
}
type transaction struct {
	Schema              string               `json:"schema"`
	TransactionID       string               `json:"transaction_id"`
	AuthorityRoutes     []authorityRoute     `json:"authority_routes"`
	RoleBindings        []roleBinding        `json:"role_bindings"`
	DecisionTransitions []decisionTransition `json:"decision_transitions"`
	SnapshotBindings    []snapshotBinding    `json:"snapshot_bindings"`
	RawReconstructions  []json.RawMessage    `json:"raw_reconstructions"`
}

type rawObservation struct {
	EvidenceGroupsObserved   int    `json:"evidence_groups_observed"`
	EvidenceGroupsTotal      int    `json:"evidence_groups_total"`
	SelfMintingPaths         *int   `json:"self_minting_paths"`
	RoleConflictPaths        *int   `json:"role_conflict_paths"`
	UnknownLaunderingPaths   *int   `json:"unknown_laundering_paths"`
	UnknownTopDecisions      *int   `json:"unknown_top_decisions"`
	SnapshotBindingsObserved int    `json:"snapshot_bindings_observed"`
	SnapshotBindingsRequired int    `json:"snapshot_bindings_required"`
	ExactSnapshotBindingBPS  *int   `json:"exact_snapshot_binding_bps"`
	SnapshotMismatchPaths    *int   `json:"snapshot_mismatch_paths"`
	CandidateDecision        string `json:"candidate_decision"`
	CandidateReason          string `json:"candidate_reason"`
	CandidateResolution      string `json:"candidate_resolution"`
}

type rawReceipt struct {
	Schema               string         `json:"schema"`
	VerifierID           string         `json:"verifier_id"`
	SubjectSHA           string         `json:"subject_sha"`
	DenominatorDigest    string         `json:"denominator_digest"`
	RawTransactionDigest string         `json:"raw_transaction_digest"`
	Observation          rawObservation `json:"observation"`
}
