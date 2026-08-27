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
	return DigestBytes([]byte(fmt.Sprintf("claim-transition/v1|%s|%s|OPEN|%s", experimentID, gateID, next)))
}

func artifactDigest(path string, bytes int) string {
	return DigestBytes([]byte(fmt.Sprintf("artifact/v1|%s|%d", path, bytes)))
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || value[:len("sha256:")] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}
