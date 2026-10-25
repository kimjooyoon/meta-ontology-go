package main

import (
	"encoding/hex"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/scripts/source-splitter/completion"
)

// This consumer checks the exact logical write set, not completed construction.
func splitCompletionKnown(report splitReport) bool {
	c := report.Coordinates
	if report.PolicySourceDigest == "" {
		return report.PlanSourceDigest == "" && report.NextOperation == "" && c.Input == 0 && c.Unhandled == 0 &&
			(report.Decision == "PASS" || report.Decision == "FIXED_POINT") && report.Resolution == "EXACT"
	}
	if report.PolicySourceDigest != completion.SourceDigest() || !splitPlanDigestKnown(report.PlanSourceDigest) ||
		c.Input != c.Selected+len(report.Unhandled) || c.Unhandled != len(report.Unhandled) {
		return false
	}
	reason := completion.Classify(int64(c.Selected), int64(c.Applied), int64(c.Deferred), int64(c.Unhandled))
	decision := completion.Decision(reason)
	return (decision == "PASS" || decision == "FIXED_POINT" || decision == "PROGRESS") &&
		report.Reason == reason && report.Decision == decision && report.Resolution == completion.Resolution(reason) &&
		report.NextOperation == completion.NextOperation(reason)
}

func splitPlanDigestKnown(digest string) bool {
	raw, ok := strings.CutPrefix(digest, "sha256:")
	if !ok || len(raw) != 64 {
		return false
	}
	_, err := hex.DecodeString(raw)
	return err == nil
}
