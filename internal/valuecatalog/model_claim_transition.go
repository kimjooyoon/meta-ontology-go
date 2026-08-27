package valuecatalog

import "strings"

const (
	ClaimStatusUnrecorded       = "UNRECORDED"
	ClaimStatusOpen             = "OPEN"
	ClaimStatusDischarged       = "DISCHARGED"
	ClaimEventRegistered        = "CLAIM_REGISTERED"
	ClaimEventEvidenceAccepted  = "EVIDENCE_ACCEPTED"
	ClaimEventEvidenceUnavailable = "EVIDENCE_UNAVAILABLE"
	ReasonClaimDeclared         = "CLAIM_DECLARED"
	ReasonClaimEvidenceAccepted = "CLAIM_EVIDENCE_ACCEPTED"
)

type ClaimTransition struct {
	Sequence                 int               `json:"sequence"`
	ClaimID                  string            `json:"claim_id"`
	DeclarationDigest        string            `json:"declaration_digest"`
	Event                    string            `json:"event"`
	Before                   string            `json:"before"`
	After                    string            `json:"after"`
	Coordinate               ProcessCoordinate `json:"coordinate"`
	EvidenceDigest           string            `json:"evidence_digest,omitempty"`
	PreviousTransitionDigest string            `json:"previous_transition_digest,omitempty"`
	TransitionDigest         string            `json:"transition_digest"`
}

func claimDeclarationDigest(claim Claim) string {
	return digestValue(struct {
		ClaimID, Stage, Statement string
	}{claim.ClaimID, claim.Stage, claim.Statement})
}

func claimTransitionDigest(transition ClaimTransition) string {
	transition.TransitionDigest = ""
	return digestValue(transition)
}

func appendClaimTransition(ledger []ClaimTransition, transition ClaimTransition) []ClaimTransition {
	transition.Sequence = len(ledger) + 1
	if len(ledger) > 0 {
		transition.PreviousTransitionDigest = ledger[len(ledger)-1].TransitionDigest
	}
	transition.TransitionDigest = claimTransitionDigest(transition)
	return append(ledger, transition)
}

func claimCoordinate(claim Claim, reason string) ProcessCoordinate {
	stage, step, _ := strings.Cut(claim.Stage, "/")
	return ProcessCoordinate{Stage: stage, Step: step, Reason: reason}
}
