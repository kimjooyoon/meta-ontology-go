package main

import (
	"strings"
	"testing"
)

func TestEnvelopeBindsTheCompleteCycle(t *testing.T) {
	opts, in := cycleFixture()
	envelope := buildEnvelope(in, opts)
	if envelope.Status != "BOUND" {
		t.Fatalf("status = %s, want BOUND", envelope.Status)
	}
	if len(envelope.Artifacts) != 6 || len(envelope.Indicators) != 9 {
		t.Fatalf("artifacts/indicators = %d/%d", len(envelope.Artifacts), len(envelope.Indicators))
	}
	if envelope.PromotionAuthorized || !validDigest(envelope.EnvelopeDigest) ||
		!validDigest(envelope.ReplayDigest) {
		t.Fatalf("invalid envelope digests or authority: %#v", envelope)
	}
}

func TestEnvelopeRejectsAPlanLinkDrift(t *testing.T) {
	opts, in := cycleFixture()
	in.Execution.Value.PlanDigest = strings.Repeat("2", 64)
	envelope := buildEnvelope(in, opts)
	if envelope.Status != "OPEN" {
		t.Fatalf("status = %s, want OPEN", envelope.Status)
	}
}

func cycleFixture() (options, inputs) {
	head, base, digest := strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("1", 64)
	ledger := "sha256:" + digest
	docHash := strings.Repeat("3", 64)
	opts := options{headSHA: head, branch: "dev", conclusion: "success", runID: 42}
	in := inputs{
		Metrics:    document[metricsDocument]{Value: metricsDocument{CommitSHA: head}, FileSHA256: docHash},
		Plan:       document[planDocument]{Value: planDocument{SchemaVersion: "gooo/self-improvement-generation/v6", BaseSHA: base, HeadSHA: head, PlanDigest: digest, Decision: "FIXED_POINT", Reason: "EXACT_FIXED_POINT"}, FileSHA256: docHash},
		Execution:  document[executionDocument]{Value: executionDocument{SchemaVersion: "gooo/meta-operation-execution/v6", BaseSHA: base, HeadSHA: head, PlanDigest: digest, ManifestDigest: digest, IndicatorDecisionLedgerDigest: ledger, IndicatorDecisionLedgerCount: 1, Decision: "FIXED_POINT", Reason: "EXACT_FIXED_POINT"}, FileSHA256: docHash},
		Receipts:   document[receiptDocument]{Value: receiptDocument{SchemaVersion: "gooo/meta-operation-receipt-report/v2", BaseSHA: base, HeadSHA: head, PlanDigest: digest, ReportDigest: digest, IndicatorDecisionLedgerDigest: ledger, IndicatorDecisionLedgerCount: 1, Decision: "FIXED_POINT", Reason: "EXACT_FIXED_POINT"}, FileSHA256: docHash},
		Provenance: document[provenanceDocument]{Value: provenanceDocument{SchemaVersion: "gooo/meta-artifact-provenance/v1", BaseSHA: base, HeadSHA: head, PlanDigest: digest, ExecutionManifestDigest: digest, ReceiptReportDigest: digest, IndicatorDecisionLedgerDigest: ledger, IndicatorDecisionLedgerCount: 1, Decision: "BOUND", Reason: "ARTIFACT_PROVENANCE_BOUND", EnvelopeDigest: digest}, FileSHA256: docHash},
		Contract:   document[contractDocument]{Value: contractFixture(head, digest), FileSHA256: docHash},
	}
	return opts, in
}

func contractFixture(head, digest string) contractDocument {
	indicators := []contractIndicator{{"FOUNDATION", "PASS"}, {"FOUNDATION", "PASS"}, {"FOUNDATION", "PASS"}, {"COHERENCE", "PASS"}, {"COHERENCE", "PASS"}, {"COHERENCE", "PASS"}, {"REGRESSION", "PASS"}}
	coverage := []contractCoverage{{true}, {true}, {true}}
	return contractDocument{Schema: "gooo/self-improvement-contract/v1", CommitSHA: head,
		SourceSHA256: digest, SemanticHash: digest, RegistryDigest: digest,
		Status: "PASS", Indicators: indicators, ExecutorCoverage: coverage}
}
