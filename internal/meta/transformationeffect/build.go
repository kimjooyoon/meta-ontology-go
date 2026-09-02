package transformationeffect

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect/workspace"
)

func Build(opts Options) (Result, error) {
	in, err := loadInputs(opts)
	if err != nil {
		return Result{}, err
	}
	sourceBefore, err := workspace.Scan(opts.Root)
	if err != nil {
		return Result{}, err
	}
	executed, err := executePlan(in, opts, sourceBefore)
	if err != nil {
		return Result{}, err
	}
	sourceAfter, err := workspace.Scan(opts.Root)
	if err != nil || sourceAfter.Digest != sourceBefore.Digest {
		return Result{}, fmt.Errorf("source workspace changed outside sandbox: %w", err)
	}
	decision, reason := "APPLIED", "SANDBOX_EFFECTS_VERIFIED"
	if in.plan.Decision == generation.DecisionFixedPoint {
		decision, reason = "FIXED_POINT", "EXACT_FIXED_POINT"
		if len(executed.effects) != 0 || len(executed.patch.Changes) != 0 || executed.baseline.Digest != executed.final.Digest {
			return Result{}, fmt.Errorf("fixed point produced an effect")
		}
	} else if !validExecutionOutcome(executed.receipts, len(in.plan.Selected)) || len(executed.effects) != len(in.plan.Selected) {
		return Result{}, fmt.Errorf("planned effects are not conformant")
	}
	causal, err := BuildCausalUnknownProjection(executed.receipts)
	if err != nil {
		return Result{}, fmt.Errorf("causal unknown projection: %w", err)
	}
	ledger := Ledger{Schema: ledgerSchema, Metaprogram: "scripts/transformation-effect",
		BaseSHA: in.plan.BaseSHA, HeadSHA: in.plan.HeadSHA, SourceSchema: in.metrics.Meta.Schema,
		RootTopologyExempt: true, Artifacts: in.digests, InputDigest: hashJSON(in.digests),
		IndicatorLedgerDigest: in.plan.IndicatorDecisionLedger.Digest,
		IndicatorLedgerCount:  in.plan.IndicatorDecisionLedger.IndicatorCount,
		Decision:              decision, Reason: reason, WorkspaceMode: string(generation.WorkspaceModeDisposable),
		WriteBoundary: string(generation.WriteBoundarySandboxOnly), SourceTreeBefore: sourceBefore.Digest,
		SourceTreeAfter: sourceAfter.Digest, SourceWorkspaceUnchanged: true,
		SandboxTreeBefore: executed.baseline.Digest, SandboxTreeAfter: executed.final.Digest,
		Effects: executed.effects, PatchDigest: executed.patch.PatchDigest,
		InputReceiptReportDigest:     in.receipts.ReportDigest,
		GeneratedReceiptReportDigest: executed.receipts.ReportDigest,
		InputProvenanceDigest:        in.provenance.EnvelopeDigest,
		ExecutedProvenanceDigest:     executed.provenance.EnvelopeDigest, Status: "BOUND",
		SelectedPlanOperations:        executed.selectedPlanOperations,
		BoundExecutorOperations:       executed.boundExecutorOperations,
		UnboundExecutorOperations:     executed.unboundExecutorOperations,
		OperationOutcome:              operationOutcome(executed.receipts),
		ReceiptDecision:               string(executed.receipts.Decision),
		ReceiptCount:                  len(executed.receipts.Receipts),
		FailureCount:                  len(executed.receipts.Failures),
		UnknownCount:                  len(executed.receipts.Unknowns),
		DirectUnknownCount:            causal.DirectUnknownCount,
		DependencyBlockedUnknownCount: causal.DependencyBlockedUnknownCount,
		UnknownCausalDigest:           causal.Digest}
	ledger.Indicators = effectIndicators(ledger, len(in.plan.Selected), executed.receipts.Decision)
	ledger = sealLedger(ledger)
	if err := validateLedger(ledger); err != nil {
		return Result{}, err
	}
	if err := validateSplitGoEffects(ledger.Effects); err != nil {
		return Result{}, err
	}
	return Result{ledger, executed.patch, executed.receipts, executed.provenance}, nil
}

func operationOutcome(report generation.ReceiptReport) string {
	if report.Decision == generation.ReceiptDecisionFixedPoint {
		return OperationOutcomeFixedPoint
	}
	if report.Decision == generation.ReceiptDecisionConformant {
		return OperationOutcomeClosed
	}
	if report.Decision == generation.ReceiptDecisionRefuted &&
		len(report.Receipts) > 0 && len(report.Failures) > 0 {
		return OperationOutcomeMixedClosedRefuted
	}
	return string(report.Decision)
}

func validExecutionOutcome(report generation.ReceiptReport, selected int) bool {
	if len(report.Receipts)+len(report.Failures) != selected {
		return false
	}
	if report.Decision == generation.ReceiptDecisionConformant {
		return len(report.Failures) == 0
	}
	if report.Decision != generation.ReceiptDecisionRefuted || len(report.Failures) == 0 {
		return false
	}
	for _, failure := range report.Failures {
		if failure.Decision != string(generation.ReceiptDecisionRefuted) {
			return false
		}
	}
	return true
}
