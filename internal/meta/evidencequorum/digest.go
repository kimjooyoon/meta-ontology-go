package evidencequorum

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func SourceDigest(raw []byte) string { return digestBytes(raw) }

func digestJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(raw)
}

func validDigest(value string) bool { return digestPattern.MatchString(value) }
func validHead(value string) bool   { return headPattern.MatchString(value) }

func SealReceipt(value Receipt) Receipt {
	value.Digest = ""
	value.Digest = digestJSON(value)
	return value
}

func verifyReceipt(value Receipt) bool {
	digest := value.Digest
	value.Digest = ""
	return validDigest(digest) && digest == digestJSON(value)
}
