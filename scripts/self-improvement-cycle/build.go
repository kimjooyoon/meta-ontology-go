package main

import (
	"bytes"
	"encoding/json"
)

const envelopeSchema, metaprogram = "gooo/self-improvement-cycle-envelope/v3", "scripts/self-improvement-cycle"

type projection struct {
	Envelope   Envelope
	Validation validation
}

func projectEnvelope(in inputs, opts options) projection {
	plan, execution := in.Plan.Value, in.Execution.Value
	receipts, provenance, contract := in.Receipts.Value, in.Provenance.Value, in.Contract.Value
	metrics := metricsProjection(in.Metrics.Value)
	artifacts := []ArtifactRef{
		{Kind: "contract", Schema: contract.Schema, FileSHA256: in.Contract.FileSHA256, SemanticDigest: contract.SemanticHash, Decision: contract.Status},
		{Kind: "source-metrics", Schema: metrics.Schema, FileSHA256: in.Metrics.FileSHA256, SemanticDigest: metrics.SemanticDigest, Decision: "OBSERVED"},
		{Kind: "generation-plan", Schema: plan.SchemaVersion, FileSHA256: in.Plan.FileSHA256, SemanticDigest: plan.PlanDigest, Decision: plan.Decision},
		{Kind: "execution-manifest", Schema: execution.SchemaVersion, FileSHA256: in.Execution.FileSHA256, SemanticDigest: execution.ManifestDigest, Decision: execution.Decision},
		{Kind: "receipt-report", Schema: receipts.SchemaVersion, FileSHA256: in.Receipts.FileSHA256, SemanticDigest: receipts.ReportDigest, Decision: receipts.Decision},
		{Kind: "artifact-provenance", Schema: provenance.SchemaVersion, FileSHA256: in.Provenance.FileSHA256, SemanticDigest: provenance.EnvelopeDigest, Decision: provenance.Decision},
	}
	artifactSetDigest := digestJSON(artifacts)
	context := struct {
		HeadSHA, Branch, Conclusion, ArtifactSetDigest string
		RunID                                          int64
	}{opts.headSHA, opts.branch, opts.conclusion, artifactSetDigest, opts.runID}
	return projection{Envelope: Envelope{
		Schema: envelopeSchema, Metaprogram: metaprogram,
		BaseSHA: plan.BaseSHA, HeadSHA: opts.headSHA,
		CIWorkflowRunID: opts.runID, CIHeadBranch: opts.branch, CIConclusion: opts.conclusion,
		ContractSemanticHash: contract.SemanticHash, ContractRegistryDigest: contract.RegistryDigest,
		SourceMetrics:         metrics,
		IndicatorLedgerDigest: provenance.IndicatorDecisionLedgerDigest,
		IndicatorLedgerCount:  provenance.IndicatorDecisionLedgerCount,
		Artifacts:             artifacts, ArtifactSetDigest: artifactSetDigest,
		InputDigest: digestJSON(context), Indicators: []Indicator{},
		PromotionAuthorized: false,
	}, Validation: validateInputs(in, opts)}
}

func buildEnvelope(in inputs, opts options) Envelope {
	left, right := projectEnvelope(in, opts), projectEnvelope(in, opts)
	leftBytes, _ := json.Marshal(left.Envelope)
	rightBytes, _ := json.Marshal(right.Envelope)
	left.Envelope.Indicators = cycleIndicators(left.Validation, left.Envelope, bytes.Equal(leftBytes, rightBytes))
	finishEnvelope(&left.Envelope)
	canonical := left.Envelope
	canonical.EnvelopeDigest, canonical.ReplayDigest = "", ""
	left.Envelope.EnvelopeDigest = digestJSON(canonical)
	left.Envelope.ReplayDigest = digestJSON(struct {
		EnvelopeDigest, ArtifactSetDigest, InputDigest string
	}{left.Envelope.EnvelopeDigest, left.Envelope.ArtifactSetDigest, left.Envelope.InputDigest})
	return left.Envelope
}

func metricsProjection(metrics metricsDocument) MetricsBinding {
	result := MetricsBinding{Schema: metrics.Meta.Schema, RootTopologyExempt: metrics.Meta.Policy.ExemptProjectRootTopology}
	for _, root := range metrics.Directories {
		if root.Path == "." {
			result.LogicalRoot = root
		}
	}
	for _, root := range metrics.StorageDirectories {
		if root.Path == "." {
			result.StorageRoot = root
		}
	}
	result.RootWitnessDigest, result.RootWitnessCount = metricWitnessBinding(metrics, result)
	result.SemanticDigest = digestJSON(result)
	return result
}
