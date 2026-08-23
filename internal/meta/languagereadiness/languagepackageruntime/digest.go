package languagepackageruntime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil { panic(err) }
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func reportDigest(report Report) string {
	report.ReportDigest = ""
	return digestValue(report)
}
