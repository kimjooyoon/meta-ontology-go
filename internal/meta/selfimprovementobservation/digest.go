package selfimprovementobservation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

var (
	shaPattern       = regexp.MustCompile("^[0-9a-f]{40}$")
	rawDigestPattern = regexp.MustCompile("^[0-9a-f]{64}$")
	digestPattern    = regexp.MustCompile("^sha256:[0-9a-f]{64}$")
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return digestBytes(data)
}

func validSourceReportDigest(report SourceReport) bool {
	provided := report.Digest
	report.Digest = ""
	return digestPattern.MatchString(provided) && provided == digestJSON(report)
}

func ValidObservationDigest(observation Observation) bool {
	provided := observation.Digest
	observation.Digest = ""
	return digestPattern.MatchString(provided) && provided == digestJSON(observation)
}
