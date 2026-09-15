package selfimprovementtransport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

var (
	shaPattern    = regexp.MustCompile(`^[0-9a-f]{40}$`)
	digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func validSHA(value string) bool { return shaPattern.MatchString(value) }

func validDigest(value string) bool { return digestPattern.MatchString(value) }
