package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	commitPattern = regexp.MustCompile("^[0-9a-f]{40}$")
	digestPattern = regexp.MustCompile("^[0-9a-f]{64}$")
)

func contractIndicators(result analysis, replay bool) []Indicator {
	report := result.Report
	covered, total := coveredCount(report.ExecutorCoverage), len(report.ExecutorCoverage)
	semanticValue := report.SemanticHash
	if semanticValue == "" {
		semanticValue = "missing"
	}
	return []Indicator{
		{
			ID: "foundation.exact-head-binding", Route: "FOUNDATION",
			Verdict: verdict(commitPattern.MatchString(report.CommitSHA)),
			Relation: "matches", Value: report.CommitSHA, Limit: "40-lower-hex",
		},
		{
			ID: "foundation.gooo-source-binding", Route: "FOUNDATION",
			Verdict: verdict(strings.HasSuffix(report.ContractPath, ".gooo") &&
				digestPattern.MatchString(report.SourceSHA256)),
			Relation: "sha256", Value: report.SourceSHA256, Limit: "bound",
		},
		{
			ID: "foundation.semantic-lowering", Route: "FOUNDATION",
			Verdict: verdict(result.SemanticOK), Relation: "stable-hash",
			Value: semanticValue, Limit: "present",
		},
		{
			ID: "coherence.closed-self-improvement-loop", Route: "COHERENCE",
			Verdict: verdict(result.LoopOK), Relation: "=", Value: fmt.Sprint(result.LoopOK),
			Limit: "true",
		},
		{
			ID: "coherence.executor-registry-coverage", Route: "COHERENCE",
			Verdict: verdict(result.ExecutorOK), Relation: "=",
			Value: fmt.Sprintf("%d/%d", covered, total),
			Limit: fmt.Sprintf("%d/%d", total, total),
		},
		{
			ID: "coherence.trilemma-choice", Route: "COHERENCE",
			Verdict: verdict(result.TrilemmaOK), Relation: "=",
			Value: fmt.Sprint(result.TrilemmaOK), Limit: "true",
		},
		{
			ID: "regression.canonical-replay", Route: "REGRESSION",
			Verdict: verdict(replay), Relation: "=",
			Value: fmt.Sprint(replay), Limit: "true",
		},
	}
}

func finishReport(report *Report) {
	report.Status, report.Reason = "PASS", "SELF_IMPROVEMENT_CONTRACT_CLOSED"
	for _, indicator := range report.Indicators {
		if indicator.Verdict != "PASS" {
			report.Status, report.Reason = "FAIL", "SELF_IMPROVEMENT_CONTRACT_OPEN"
			return
		}
	}
}

func verdict(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}
