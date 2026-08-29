package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
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
		switch id {
		case "metrics":
			writeFixture(t, path, sourceMetricsFixture(t))
		case "plan":
			writeFixture(t, path, sourcePlanFixture(t))
		case "provenance":
			continue
		default:
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

func sourceMetricsFixture(t *testing.T) []byte {
	t.Helper()
	meta := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		{MetricID: sourcepolicy.DimensionRootREADME, Family: sourcepolicy.FamilyDocumentation, Subject: ".", SubjectKind: sourcepolicy.SubjectKindProjectRoot, Applicability: sourcepolicy.ApplicabilityNotApplicable, ApplicabilityRule: sourcepolicy.ApplicabilityRuleProjectRootREADME, ApplicabilityReason: sourcepolicy.ApplicabilityReasonRootREADMEExempt, Satisfied: true, Proof: sourcepolicy.ProofFoundation, Producer: "fixture", Consumer: "metric-meta-program", Operation: sourcepolicy.OperationExemptRootREADME},
		{MetricID: sourcepolicy.DimensionGoFileLines, Family: sourcepolicy.FamilyVolume, Subject: "fixture.go", SubjectKind: sourcepolicy.SubjectKindFile, Value: 76, Limit: 75, Relation: sourcepolicy.RelationLessOrEqual, Applicability: sourcepolicy.ApplicabilityApplicable, ApplicabilityRule: sourcepolicy.ApplicabilityRuleDefault, ApplicabilityReason: sourcepolicy.ApplicabilityReasonCatalogApplicable, Blocking: false, Satisfied: false, Role: sourcepolicy.IndicatorRoleDriver, Proof: sourcepolicy.ProofFoundation, Producer: "fixture", Consumer: "source-splitter", Operation: sourcepolicy.OperationSplitGo},
	}}
	document := map[string]any{
		"root": ".", "files": []map[string]any{{"path": "fixture.go", "language": "go", "lines": 76}},
		"directories": []map[string]any{{"path": ".", "recursive_files": 1, "recursive_folders": 0, "go_files": 1, "go_lines": 76, "gooo_files": 0, "gooo_lines": 0}},
		"meta":        meta,
	}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func sourcePlanFixture(t *testing.T) []byte {
	t.Helper()
	document := map[string]any{
		"schema_version": "gooo/self-improvement-generation/v6",
		"selected_count": 2,
		"selected": []map[string]any{
			{"meta_operation": "split-go-declarations", "metric_id": string(sourcepolicy.DimensionGoFileLines), "subject": "fixture.go"},
			{"meta_operation": "extract-function", "metric_id": string(sourcepolicy.DimensionFunctionLines), "subject": "fixture.go:1:Fixture"},
		},
	}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
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
