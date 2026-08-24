package languagedelivery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(data)
}

func factsDigest(report Report) string {
	return digestValue(struct {
		ContractDigest string
		ManifestDigest string
		Summary Summary
		Sources []SourceObservation
		Obligations []ObligationResult
	}{report.ContractDigest, report.ManifestDigest, report.Summary, report.Sources, report.Obligations})
}
