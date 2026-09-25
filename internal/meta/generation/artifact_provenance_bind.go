package generation

import "strconv"

func BindArtifactProvenance(
	plan Plan,
	execution ExecutionManifest,
	receipts ReceiptReport,
) ArtifactProvenance {
	ledgerDigest, ledgerCount := planIndicatorDecisionLedgerProvenance(plan)
	planCanonical := artifactCanonical(plan)
	executionCanonical := artifactCanonical(execution)
	receiptsCanonical := artifactCanonical(receipts)
	envelope := ArtifactProvenance{
		SchemaVersion: ArtifactProvenanceSchemaVersion,
		BaseSHA:       plan.BaseSHA, HeadSHA: plan.HeadSHA,
		PlanDigest:                    plan.PlanDigest,
		ExecutionManifestDigest:       execution.ManifestDigest,
		ReceiptReportDigest:           receipts.ReportDigest,
		IndicatorDecisionLedgerDigest: ledgerDigest,
		IndicatorDecisionLedgerCount:  ledgerCount,
		Indicators: []ArtifactProvenanceIndicator{
			artifactProvenanceIndicator("foundation.plan-ledger", TrilemmaRouteFoundation,
				artifactBindingVerdict(planCanonical, ledgerDigest != ""),
				plan.PlanDigest, ledgerDigest, strconv.Itoa(ledgerCount)),
			artifactProvenanceIndicator("coherence.execution-ledger", TrilemmaRouteCoherence,
				artifactBindingVerdict(executionCanonical,
					executionMatchesProvenance(plan, execution, ledgerDigest, ledgerCount)),
				execution.PlanDigest, execution.IndicatorDecisionLedgerDigest),
			artifactProvenanceIndicator("coherence.receipt-ledger", TrilemmaRouteCoherence,
				artifactBindingVerdict(receiptsCanonical,
					receiptsMatchProvenance(plan, receipts, ledgerDigest, ledgerCount)),
				receipts.PlanDigest, receipts.IndicatorDecisionLedgerDigest),
			artifactProvenanceIndicator("regression.canonical-replay", TrilemmaRouteRegression,
				artifactBindingVerdict(planCanonical && executionCanonical && receiptsCanonical, true),
				plan.PlanDigest, execution.ReplayDigest, receipts.ReplayDigest),
		}}
	return finishArtifactProvenance(envelope)
}
