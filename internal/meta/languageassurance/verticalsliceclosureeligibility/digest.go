package verticalsliceclosureeligibility

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(encoded)
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func binding(capsule Capsule) ArtifactBinding {
	observed := digestBytes(capsule.Payload)
	return ArtifactBinding{Name: capsule.Name, ArtifactID: capsule.ArtifactID,
		ArchiveDigest: capsule.ArchiveDigest, CapsuleDigest: capsule.CapsuleDigest,
		ObservedDigest: observed, Exact: capsuleExact(capsule)}
}

func capsuleExact(capsule Capsule) bool {
	switch capsule.Name {
	case AssuranceName:
		return capsule.ArtifactID == AssuranceArtifactID &&
			capsule.ArchiveDigest == AssuranceArchiveDigest &&
			capsule.CapsuleDigest == AssuranceCapsuleDigest &&
			digestBytes(capsule.Payload) == AssuranceCapsuleDigest
	case ShadowName:
		return capsule.ArtifactID == ShadowArtifactID &&
			capsule.ArchiveDigest == ShadowArchiveDigest &&
			capsule.CapsuleDigest == ShadowCapsuleDigest &&
			digestBytes(capsule.Payload) == ShadowCapsuleDigest
	default:
		return false
	}
}

func bindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{binding(input.Assurance), binding(input.Shadow)}
}

func countExact(values []ArtifactBinding) int {
	count := 0
	for _, value := range values {
		if value.Exact {
			count++
		}
	}
	return count
}

func sealReport(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}

func sealSuite(suite Suite) Suite {
	suite.SuiteDigest = ""
	suite.SuiteDigest = digestJSON(suite)
	return suite
}
