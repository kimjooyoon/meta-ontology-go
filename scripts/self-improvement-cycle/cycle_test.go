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
	if len(envelope.Artifacts) != 6 || len(envelope.Indicators) != 15 {
		t.Fatalf("artifacts/indicators = %d/%d", len(envelope.Artifacts), len(envelope.Indicators))
	}
	if !envelope.SourceMetrics.RootTopologyExempt || envelope.SourceMetrics.RootWitnessCount != 10 {
		t.Fatalf("root witness = %#v", envelope.SourceMetrics)
	}
	if envelope.PromotionAuthorized || !validDigest(envelope.EnvelopeDigest) || !validDigest(envelope.ReplayDigest) {
		t.Fatalf("invalid envelope digests or authority: %#v", envelope)
	}
}

func TestEnvelopeRejectsAPlanLinkDrift(t *testing.T) {
	opts, in := cycleFixture()
	in.Execution.Value.PlanDigest = strings.Repeat("2", 64)
	if status := buildEnvelope(in, opts).Status; status != "OPEN" {
		t.Fatalf("status = %s, want OPEN", status)
	}
}

func TestEnvelopeRejectsRootExceptionDrift(t *testing.T) {
	opts, in := cycleFixture()
	before := buildEnvelope(in, opts)
	in.Metrics.Value.Meta.Policy.ExemptProjectRootTopology = false
	after := buildEnvelope(in, opts)
	if after.Status != "OPEN" || after.SourceMetrics.RootWitnessDigest == before.SourceMetrics.RootWitnessDigest {
		t.Fatalf("root exception drift was not content-addressed: %#v", after.SourceMetrics)
	}
}

func cycleFixture() (options, inputs) {
	head, base, digest := strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("1", 64)
	ledger, docHash := "sha256:"+digest, strings.Repeat("3", 64)
	opts := options{headSHA: head, branch: "dev", conclusion: "success", runID: 42}
	in := inputs{
		Metrics:    document[metricsDocument]{Value: metricsFixture(head), FileSHA256: docHash},
		Plan:       document[planDocument]{Value: planDocument{SchemaVersion: "gooo/self-improvement-generation/v6", BaseSHA: base, HeadSHA: head, PlanDigest: digest, Decision: "FIXED_POINT", Reason: "EXACT_FIXED_POINT"}, FileSHA256: docHash},
		Execution:  document[executionDocument]{Value: executionDocument{SchemaVersion: "gooo/meta-operation-execution/v6", BaseSHA: base, HeadSHA: head, PlanDigest: digest, ManifestDigest: digest, IndicatorDecisionLedgerDigest: ledger, IndicatorDecisionLedgerCount: 1, Decision: "FIXED_POINT", Reason: "EXACT_FIXED_POINT"}, FileSHA256: docHash},
		Receipts:   document[receiptDocument]{Value: receiptDocument{SchemaVersion: "gooo/meta-operation-receipt-report/v2", BaseSHA: base, HeadSHA: head, PlanDigest: digest, ReportDigest: digest, IndicatorDecisionLedgerDigest: ledger, IndicatorDecisionLedgerCount: 1, Decision: "FIXED_POINT", Reason: "EXACT_FIXED_POINT"}, FileSHA256: docHash},
		Provenance: document[provenanceDocument]{Value: provenanceDocument{SchemaVersion: "gooo/meta-artifact-provenance/v1", BaseSHA: base, HeadSHA: head, PlanDigest: digest, ExecutionManifestDigest: digest, ReceiptReportDigest: digest, IndicatorDecisionLedgerDigest: ledger, IndicatorDecisionLedgerCount: 1, Decision: "BOUND", Reason: "ARTIFACT_PROVENANCE_BOUND", EnvelopeDigest: digest}, FileSHA256: docHash},
		Contract:   document[contractDocument]{Value: contractFixture(head, digest), FileSHA256: docHash},
	}
	return opts, in
}

func metricsFixture(head string) metricsDocument {
	logical := MetricsSnapshot{Path: ".", SubjectKind: "PROJECT_ROOT", DirectFolders: 7, DirectFiles: 7, RecursiveFolders: 115, RecursiveFiles: 2802, GoFiles: 2689, GoooFiles: 5, GoLines: 144046, GoooLines: 56}
	storage := MetricsSnapshot{Path: ".", SubjectKind: "PROJECT_ROOT", DirectFolders: 4, DirectFiles: 1, RecursiveFolders: 684, RecursiveFiles: 2796, GoFiles: 73, GoooFiles: 2616, GoLines: 4028, GoooLines: 140033}
	binding := MetricsBinding{LogicalRoot: logical, StorageRoot: storage}
	indicators := []metricsIndicator{{MetricID: "gooo.metric.layout.direct-entries.v1", Subject: ".", Value: 5, Applicability: "NOT_APPLICABLE", ApplicabilityReason: "ROOT_TOPOLOGY_EXEMPT", Decision: "NOT_APPLICABLE"}, {MetricID: "gooo.metric.layout.entry-kinds.v1", Subject: ".", Value: 2, Applicability: "NOT_APPLICABLE", ApplicabilityReason: "ROOT_TOPOLOGY_EXEMPT", Decision: "NOT_APPLICABLE"}}
	for id, value := range metricExpectations(binding) {
		indicators = append(indicators, metricsIndicator{MetricID: id, Subject: ".", Value: value, Applicability: "APPLICABLE", Decision: "PASS"})
	}
	meta := metricsMeta{Schema: "gooo/indicator-report/v3", Policy: metricsPolicy{ExemptProjectRootTopology: true}, Indicators: indicators}
	return metricsDocument{CommitSHA: head, Meta: meta, Directories: []MetricsSnapshot{logical}, StorageDirectories: []MetricsSnapshot{storage}}
}

func contractFixture(head, digest string) contractDocument {
	indicators := []contractIndicator{{"FOUNDATION", "PASS"}, {"FOUNDATION", "PASS"}, {"FOUNDATION", "PASS"}, {"COHERENCE", "PASS"}, {"COHERENCE", "PASS"}, {"COHERENCE", "PASS"}, {"REGRESSION", "PASS"}}
	coverage := []contractCoverage{{true}, {true}, {true}}
	return contractDocument{Schema: "gooo/self-improvement-contract/v1", CommitSHA: head, SourceSHA256: digest, SemanticHash: digest, RegistryDigest: digest, Status: "PASS", Indicators: indicators, ExecutorCoverage: coverage}
}
