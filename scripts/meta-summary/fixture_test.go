package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureOptions(t *testing.T, decision string) options {
	t.Helper()
	dir := t.TempDir()
	paths := map[string]string{
		"metrics":    filepath.Join(dir, "source-metrics.json"),
		"plan":       filepath.Join(dir, "self-improvement-plan.json"),
		"execution":  filepath.Join(dir, "self-improvement-execution.json"),
		"receipts":   filepath.Join(dir, "self-improvement-receipts.json"),
		"provenance": filepath.Join(dir, "self-improvement-provenance.json"),
	}
	for id, path := range paths {
		if id != "provenance" {
			writeFixture(t, path, []byte("{}\n"))
		}
	}
	provenance := provenanceEnvelope{
		SchemaVersion: "gooo/meta-artifact-provenance/v1",
		BaseSHA:       strings.Repeat("a", 40), HeadSHA: strings.Repeat("b", 40),
		PlanDigest: digestFixture("a"), ExecutionManifestDigest: digestFixture("b"),
		ReceiptReportDigest: digestFixture("c"), InputDigest: digestFixture("d"),
		IndicatorDecisionLedgerDigest: "sha256:" + digestFixture("e"),
		IndicatorDecisionLedgerCount:  4, Decision: decision, Reason: "ARTIFACT_PROVENANCE_BOUND",
		Indicators: []provenanceIndicator{
			{ID: "a", Route: "FOUNDATION", Verdict: "PASS", EvidenceDigest: digestFixture("1")},
			{ID: "b", Route: "COHERENCE", Verdict: "PASS", EvidenceDigest: digestFixture("2")},
			{ID: "c", Route: "COHERENCE", Verdict: "PASS", EvidenceDigest: digestFixture("3")},
			{ID: "d", Route: "REGRESSION", Verdict: "PASS", EvidenceDigest: digestFixture("4")},
		},
		Summary: provenanceSummary{Pass: 4}, EnvelopeDigest: digestFixture("f"), ReplayDigest: digestFixture("0"),
	}
	data, err := json.Marshal(provenance)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, paths["provenance"], append(data, '\n'))
	return options{
		MetricsPath: paths["metrics"], PlanPath: paths["plan"], ExecutionPath: paths["execution"],
		ReceiptsPath: paths["receipts"], ProvenancePath: paths["provenance"],
		OutputPath: filepath.Join(dir, "summary.md"), ReportPath: filepath.Join(dir, "summary.json"),
		LimitBytes: defaultLimitBytes,
	}
}

func digestFixture(character string) string {
	return strings.Repeat(character, 64)
}

func writeFixture(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
