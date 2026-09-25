package toolchainusecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func registryDigest() string { return digestJSON(expectedRegistry()) }

func caseDigest(item CaseResult, artifactDigest string) string {
	item.EvidenceDigest = ""
	return digestJSON(struct {
		Case           CaseResult `json:"case"`
		ArtifactDigest string     `json:"artifact_digest"`
	}{item, artifactDigest})
}

func seal(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}
