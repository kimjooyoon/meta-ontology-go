package broker

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	Schema        = "gooo/capability-scoped-expansion/broker/v1"
	DecisionAllow = "ALLOW"
	DecisionDeny  = "DENY"
)

type Policy struct {
	ID                string
	DefaultDecision   string
	AuthorizationMode string
	Effects           string
}

type Request struct {
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
	PolicyID  string `json:"policy_id"`
}

type Token struct {
	id         string
	capability Request
}

func (token Token) Valid() bool {
	return token.id != ""
}

func (token Token) Capability() Request {
	return token.capability
}

type Issuance struct {
	Schema        string  `json:"schema"`
	Request       Request `json:"request"`
	RequestDigest string  `json:"request_digest"`
	PolicyID      string  `json:"policy_id"`
	PolicyDigest  string  `json:"policy_digest"`
	Decision      string  `json:"decision"`
	Issued        bool    `json:"issued"`
	Reason        string  `json:"reason"`
}

func IssueToken(policy Policy, request Request) (Token, Issuance) {
	policyDigest := digest(policy)
	requestDigest := digest(request)
	decision := DecisionDeny
	reason := "CAPABILITY_TOKEN_NOT_ISSUED_BY_DEFAULT_DENY"
	if policy.DefaultDecision == DecisionAllow && policy.AuthorizationMode == "exact-current" && policy.Effects == "NONE" && request.PolicyID == policy.ID {
		decision = DecisionAllow
		reason = "CAPABILITY_TOKEN_ISSUED"
	}
	issuance := Issuance{Schema: Schema, Request: request, RequestDigest: requestDigest, PolicyID: policy.ID, PolicyDigest: policyDigest, Decision: decision, Issued: decision == DecisionAllow, Reason: reason}
	if !issuance.Issued {
		return Token{}, issuance
	}
	return Token{id: requestDigest, capability: request}, issuance
}

func digest(value any) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
