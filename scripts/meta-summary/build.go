package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

func build(opts options) ([]byte, summaryReport, error) {
	var report summaryReport
	if opts.LimitBytes < 1 {
		return nil, report, fmt.Errorf("summary byte limit must be positive")
	}
	artifacts, provenance, err := loadArtifacts(opts)
	if err != nil {
		return nil, report, err
	}
	report = newReport(opts.LimitBytes, artifacts, provenance)
	summary, err := stabilizeSummary(&report)
	if err != nil {
		return nil, report, err
	}
	if len(summary) > opts.LimitBytes {
		report.Decision = "FAIL"
		report.Reason = "SUMMARY_BUDGET_EXCEEDED"
		report.Indicators[len(report.Indicators)-1].Verdict = "FAIL"
		return nil, report, fmt.Errorf("summary uses %d bytes; limit is %d", len(summary), opts.LimitBytes)
	}
	sum := sha256.Sum256(summary)
	report.OutputSHA256 = hex.EncodeToString(sum[:])
	return summary, report, nil
}

func newReport(limit int, artifacts []artifactEvidence, provenance provenanceEnvelope) summaryReport {
	input := digestArtifacts(artifacts)
	report := summaryReport{
		SchemaVersion: summarySchema, Decision: "PASS", Reason: "SUMMARY_WITHIN_BUDGET",
		InputSHA256: input, LimitBytes: limit, ProvenanceSchema: provenance.SchemaVersion,
		Provenance: provenanceEvidence{
			Decision: provenance.Decision, Reason: provenance.Reason,
			LedgerDigest: provenance.IndicatorDecisionLedgerDigest,
			LedgerCount:  provenance.IndicatorDecisionLedgerCount, Pass: provenance.Summary.Pass,
			Envelope: provenance.EnvelopeDigest, Replay: provenance.ReplayDigest,
		},
		Artifacts: artifacts,
		Indicators: []metricIndicator{
			{ID: "foundation.artifact-coverage", Route: "FOUNDATION", Verdict: "PASS", Relation: "=", Value: strconv.Itoa(len(artifacts)), Limit: "5"},
			{ID: "coherence.provenance-binding", Route: "COHERENCE", Verdict: "PASS", Relation: "=", Value: provenance.Decision, Limit: "BOUND"},
			{ID: "regression.canonical-projection", Route: "REGRESSION", Verdict: "PASS", Relation: "sha256", Value: input, Limit: "replay-equal"},
			{ID: "coherence.summary-bytes", Route: "COHERENCE", Verdict: "PASS", Relation: "<=", Value: "0", Limit: strconv.Itoa(limit)},
		}}
	return report
}

func stabilizeSummary(report *summaryReport) ([]byte, error) {
	for range 16 {
		report.Indicators[len(report.Indicators)-1].Value = strconv.Itoa(report.OutputBytes)
		summary := renderSummary(*report)
		if len(summary) == report.OutputBytes {
			return summary, nil
		}
		report.OutputBytes = len(summary)
	}
	return nil, fmt.Errorf("summary byte metric did not reach a fixed point")
}
