package languageproofartifactverifier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(raw)
}

func evidenceDigest(evidence Evidence) string {
	evidence.EvidenceDigest = ""
	return digestValue(evidence)
}

func artifactDigest(artifact Artifact) string {
	artifact.Digest = ""
	return digestValue(artifact)
}

func claimStatementDigest(claim ClaimStatement) string {
	claim.Digest = ""
	return digestValue(claim)
}

func claimStateDigest(claim ClaimResult) string {
	claim.StateDigest = ""
	return digestValue(claim)
}

type caseEnvelope struct {
	ID                 string     `json:"id"`
	Status             string     `json:"status"`
	ExpectedDecision   string     `json:"expected_decision"`
	ExpectedResolution string     `json:"expected_resolution"`
	ExpectedReason     string     `json:"expected_reason"`
	ObservedDecision   string     `json:"observed_decision"`
	ObservedResolution string     `json:"observed_resolution"`
	ObservedReason     string     `json:"observed_reason"`
	ProofChoice        string     `json:"proof_choice"`
	MetaOperation      string     `json:"meta_operation"`
	Coordinate         Coordinate `json:"coordinate"`
}

func caseEnvelopeValue(result CaseResult) caseEnvelope {
	return caseEnvelope{ID: result.ID, Status: result.Status, ExpectedDecision: result.ExpectedDecision,
		ExpectedResolution: result.ExpectedResolution, ExpectedReason: result.ExpectedReason,
		ObservedDecision: result.ObservedDecision, ObservedResolution: result.ObservedResolution,
		ObservedReason: result.ObservedReason, ProofChoice: result.ProofChoice, MetaOperation: result.MetaOperation,
		Coordinate: result.Coordinate}
}

func caseEnvelopeDigest(result CaseResult) string {
	return digestValue(caseEnvelopeValue(result))
}

func priorClaimStateDigest(claim ClaimResult, state string) string {
	return digestValue(struct {
		ID           string   `json:"id"`
		Proposition  string   `json:"proposition"`
		TargetDigest string   `json:"target_digest"`
		Dependencies []string `json:"dependencies"`
		State        string   `json:"state"`
	}{claim.ID, claim.Proposition, claim.TargetDigest, claim.Dependencies, state})
}

func transitionDigest(transition ClaimTransition) string {
	transition.Digest = ""
	return digestValue(transition)
}

func validDigest(value string) bool { return digestPattern.MatchString(value) }
func validHead(value string) bool   { return headPattern.MatchString(value) }
