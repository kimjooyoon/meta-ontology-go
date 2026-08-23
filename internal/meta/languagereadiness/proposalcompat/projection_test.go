package proposalcompat

import "testing"

func TestCompatibilityReceiptBindsLegacyProjection(t *testing.T) {
	head := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	legacy := sealLegacy(LegacyReceipt{Schema: LegacySchema, CurrentHeadSHA: head,
		Decision: DecisionPass, Summary: LegacySummary{Satisfied: 8, Total: 8}})
	source := Source{ExpectedHeadSHA: head,
		SourceSchema: "gooo/autonomous-change-proposal-promotion/v2",
		SourceDecision: DecisionPass, SourceReportDigest: digestJSON("source-report"),
		SourceFileSHA256: digestJSON("source-file"), SourceSatisfied: 8, SourceTotal: 8,
		TargetSchema: LegacySchema, TargetReportDigest: legacy.ReportDigest,
		TargetFileSHA256: digestBytes(EncodeLegacy(legacy)), ProjectedFields: projectedFields}
	bundle := Bundle{Legacy: legacy, Receipt: buildReceipt(source)}
	if err := Validate(bundle, head); err != nil {
		t.Fatal(err)
	}
	if bundle.Receipt.Summary.Satisfied != 6 || bundle.Receipt.Summary.FieldLosses != 0 ||
		len(bundle.Receipt.Indicators) != 8 || len(bundle.Receipt.Proofs) != 3 {
		t.Fatalf("receipt = %#v", bundle.Receipt)
	}
}

func TestCompatibilityReceiptRejectsFieldLoss(t *testing.T) {
	source := Source{SourceSchema: "gooo/autonomous-change-proposal-promotion/v2",
		SourceDecision: DecisionPass, SourceSatisfied: 8, SourceTotal: 8,
		SourceReportDigest: digestJSON("source"), SourceFileSHA256: digestJSON("file"),
		TargetSchema: LegacySchema, TargetReportDigest: digestJSON("target"),
		TargetFileSHA256: digestJSON("target-file"), ProjectedFields: projectedFields,
		FieldLosses: 1}
	report := buildReceipt(source)
	if report.Decision != DecisionFailClosed || report.Summary.NotSatisfied != 1 {
		t.Fatalf("report = %#v", report)
	}
}
