package partialknowledgecomposition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func receiptDigest(receipt Receipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

func transitionDigest(transition ClaimTransition) string {
	transition.Digest = ""
	return digestValue(transition)
}

func semanticProjectionDigest(receipt Receipt) string {
	return digestValue(struct {
		SemanticIRDigest string            `json:"semantic_ir_digest"`
		Cases            []CaseResult      `json:"cases"`
		Claims           []ClaimTransition `json:"claims"`
		Summary          Summary           `json:"summary"`
	}{receipt.SemanticIRDigest, receipt.Cases, receipt.Claims, receipt.Summary})
}
