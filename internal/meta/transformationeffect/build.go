package transformationeffect

import (
	"fmt"
	"strconv"

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
	} else if executed.receipts.Decision != generation.ReceiptDecisionConformant || len(executed.effects) != len(in.plan.Selected) {
		return Result{}, fmt.Errorf("planned effects are not conformant")
	}
	ledger := Ledger{Schema: ledgerSchema, Metaprogram: "scripts/transformation-effect",
		BaseSHA: in.plan.BaseSHA, HeadSHA: in.plan.HeadSHA, SourceSchema: in.metrics.Meta.Schema,
		RootTopologyExempt: true, Artifacts: in.digests, InputDigest: hashJSON(in.digests),
		IndicatorLedgerDigest: in.plan.IndicatorDecisionLedger.Digest, IndicatorLedgerCount: in.plan.IndicatorDecisionLedger.IndicatorCount,
		Decision:              decision, Reason: reason, WorkspaceMode: string(generation.WorkspaceModeDisposable),
		WriteBoundary: string(generation.WriteBoundarySandboxOnly), SourceTreeBefore: sourceBefore.Digest,
		SourceTreeAfter: sourceAfter.Digest, SourceWorkspaceUnchanged: true,
		SandboxTreeBefore: executed.baseline.Digest, SandboxTreeAfter: executed.final.Digest,
		Effects: executed.effects, PatchDigest: executed.patch.PatchDigest,
		InputReceiptReportDigest:     in.receipts.ReportDigest,
		GeneratedReceiptReportDigest: executed.receipts.ReportDigest,
		InputProvenanceDigest:        in.provenance.EnvelopeDigest,
		ExecutedProvenanceDigest:     executed.provenance.EnvelopeDigest, Status: "BOUND"}
	ledger.Indicators = effectIndicators(ledger, len(in.plan.Selected), executed.receipts.Decision)
	ledger = sealLedger(ledger)
	if err := validateLedger(ledger); err != nil {
		return Result{}, err
	}
	return Result{ledger, executed.patch, executed.receipts, executed.provenance}, nil
}

func effectIndicators(ledger Ledger, selected int, receipt generation.ReceiptDecision) []Indicator {
	pass := func(id, route, relation, value, limit string) Indicator {
		return Indicator{id, route, "PASS", relation, value, limit}
	}
	return []Indicator{
		pass("foundation.artifact-schemas", "FOUNDATION", "=", "true", "true"),
		pass("foundation.exact-head", "FOUNDATION", "=", ledger.HeadSHA, ledger.HeadSHA),
		pass("foundation.root-topology-exemption", "FOUNDATION", "=", "true", "true"),
		pass("foundation.indicator-ledger", "FOUNDATION", "sha256", ledger.IndicatorLedgerDigest, "bound"),
		pass("foundation.disposable-write-boundary", "FOUNDATION", "=", ledger.WriteBoundary, "SANDBOX_ONLY"),
		pass("coherence.selected-effects", "COHERENCE", "=", strconv.Itoa(len(ledger.Effects)), strconv.Itoa(selected)),
		pass("coherence.generated-receipts", "COHERENCE", "=", string(receipt), string(receipt)),
		pass("coherence.content-patch", "COHERENCE", "sha256", ledger.PatchDigest, "bound"),
		pass("coherence.executed-provenance", "COHERENCE", "sha256", ledger.ExecutedProvenanceDigest, "bound"),
		pass("regression.source-workspace", "REGRESSION", "=", "unchanged", "unchanged"),
		pass("regression.canonical-encoding", "REGRESSION", "=", "true", "true"),
	}
}
