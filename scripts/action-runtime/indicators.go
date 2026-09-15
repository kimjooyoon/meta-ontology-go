package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func buildReport(workflow string, source []byte, commit string) Report {
	left := analyze(workflow, source, commit)
	right := analyze(workflow, source, commit)
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	left.Indicators = indicators(left, bytes.Equal(leftBytes, rightBytes))
	left.Status, left.Reason = "PASS", "ACTION_RUNTIME_CONFORMANT"
	for _, indicator := range left.Indicators {
		if indicator.Verdict != "PASS" {
			left.Status, left.Reason = "FAIL", "ACTION_RUNTIME_NONCONFORMANT"
			break
		}
	}
	return left
}

func indicators(report Report, replayEqual bool) []Indicator {
	total := fmt.Sprint(report.ActionsTotal)
	return []Indicator{
		{
			ID: "foundation.catalog-coverage", Route: "FOUNDATION",
			Verdict: verdict(report.ActionsTotal > 0 &&
				report.ActionsKnown == report.ActionsTotal),
			Relation: "=", Value: fmt.Sprint(report.ActionsKnown), Limit: total,
		},
		{
			ID: "foundation.exact-head-binding", Route: "FOUNDATION",
			Verdict:  verdict(commitPattern.MatchString(report.CommitSHA)),
			Relation: "matches", Value: report.CommitSHA, Limit: "40-lower-hex",
		},
		{
			ID: "coherence.node24-runtime", Route: "COHERENCE",
			Verdict: verdict(report.ActionsTotal > 0 &&
				report.ActionsCompliant == report.ActionsTotal),
			Relation: "=", Value: fmt.Sprint(report.ActionsCompliant), Limit: total,
		},
		{
			ID: "coherence.action-input-schema", Route: "COHERENCE",
			Verdict:  verdict(report.InvalidInputsTotal == 0),
			Relation: "=", Value: fmt.Sprint(report.InvalidInputsTotal), Limit: "0",
		},
		{
			ID: "regression.canonical-replay", Route: "REGRESSION",
			Verdict: verdict(replayEqual), Relation: "=",
			Value: fmt.Sprint(replayEqual), Limit: "true",
		},
	}
}

func verdict(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}
