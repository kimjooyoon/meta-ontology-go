package experimentpromotion

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func DigestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return DigestBytes(raw)
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestValue(report)
}

func ledgerDigest(entry ClaimLedgerEntry) string {
	entry.Digest = ""
	return digestValue(entry)
}

func claimTransitionDigest(experimentID, gateID, next string) string {
	return DigestBytes([]byte(fmt.Sprintf("claim-transition/v2|%s|%s|OPEN|%s", experimentID, gateID, next)))
}

func decisionDigest(experimentID, gateID, next string) string {
	return DigestBytes([]byte(fmt.Sprintf("decision/v2|%s|%s|%s", experimentID, gateID, next)))
}

func contractedOutputDigest(semanticSourceDigest, commentSourceDigest string) string {
	return DigestBytes([]byte(fmt.Sprintf("contracted-output/v2|%s|%s", semanticSourceDigest, commentSourceDigest)))
}

func artifactDigest(raw []byte) string { return DigestBytes(raw) }

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || value[:len("sha256:")] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}
