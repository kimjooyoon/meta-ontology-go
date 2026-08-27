package causalityconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func receiptDigest(receipt Receipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

func receiptFactsDigest(receipt Receipt) string {
	return digestValue(struct {
		CandidateID      string
		SourceDigest     string
		SemanticValue    string
		Decision         string
		ClaimTransitions []ClaimTransition
		CoordinateVector []Coordinate
		Counterexamples  []Counterexample
		UnknownLocations []UnknownLocation
	}{receipt.CandidateID, receipt.SourceDigest, receipt.SemanticValue, receipt.Decision, receipt.ClaimTransitions, receipt.CoordinateVector, receipt.Counterexamples, receipt.UnknownLocations})
}

func sealCausalityReport(report CausalityReport) CausalityReport {
	report.Digest = ""
	report.Digest = digestValue(report)
	return report
}
