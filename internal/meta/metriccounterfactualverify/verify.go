package metriccounterfactualverify

import (
	"fmt"
	"os"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
)

func Replay(ledger metric.Ledger) (Receipt, error) {
	if ledger.Schema != metric.LedgerSchema || !metric.ValidLedger(ledger) {
		return Receipt{}, fmt.Errorf("invalid ledger")
	}
	if !metric.ValidManifest(ledger.Manifest) || !metric.ValidPlan(ledger.Plan) ||
		!metric.ValidState(ledger.Before) || !metric.ValidState(ledger.After) {
		return Receipt{}, fmt.Errorf("invalid sealed input")
	}
	if ledger.PromotionAuthorized || ledger.RepositoryWorkspaceWrites ||
		ledger.ExecutionPolicy != "DISPOSABLE_TEMP_ROOT_ONLY" {
		return Receipt{}, fmt.Errorf("unsafe ledger policy")
	}
	receiptSetDigest, err := metric.Digest(ledger.Receipts)
	if err != nil {
		return Receipt{}, err
	}
	if ledger.Evidence.ManifestDigest != ledger.Manifest.Digest ||
		ledger.Evidence.PlanDigest != ledger.Plan.Digest ||
		ledger.Evidence.ReceiptSetDigest != receiptSetDigest ||
		ledger.Evidence.BeforeDigest != ledger.Before.Digest ||
		ledger.Evidence.AfterDigest != ledger.After.Digest {
		return Receipt{}, fmt.Errorf("unbound ledger evidence")
	}
	root, err := os.MkdirTemp("", "gooo-metric-counterfactual-replay-")
	if err != nil {
		return Receipt{}, err
	}
	defer os.RemoveAll(root)
	if err := metric.Materialize(root, ledger.Manifest); err != nil {
		return Receipt{}, err
	}
	before, err := metric.Measure(root)
	if err != nil {
		return Receipt{}, err
	}
	receipts, err := metric.ApplyPlan(root, ledger.Plan)
	if err != nil {
		return Receipt{}, err
	}
	after, err := metric.Measure(root)
	if err != nil {
		return Receipt{}, err
	}
	delta := metric.ComputeDelta(before, after)
	indicators, err := metric.EvaluateIndicators(
		ledger.Manifest, ledger.Plan, before, after, receipts, delta,
	)
	if err != nil {
		return Receipt{}, err
	}
	if !metric.CanonicalEqual(before, ledger.Before) ||
		!metric.CanonicalEqual(after, ledger.After) ||
		!metric.CanonicalEqual(receipts, ledger.Receipts) ||
		!metric.CanonicalEqual(delta, ledger.Delta) ||
		!metric.CanonicalEqual(indicators, ledger.Indicators) ||
		!metric.AllSatisfied(indicators) {
		return Receipt{}, fmt.Errorf("counterfactual replay diverged")
	}
	replayDigest, err := metric.Digest(struct {
		Before     metric.State       `json:"before"`
		After      metric.State       `json:"after"`
		Receipts   []metric.Receipt   `json:"receipts"`
		Delta      metric.Delta       `json:"delta"`
		Indicators []metric.Indicator `json:"indicators"`
	}{before, after, receipts, delta, indicators})
	if err != nil {
		return Receipt{}, err
	}
	result := Receipt{
		Schema: Schema, LedgerDigest: ledger.Digest, ReplayDigest: replayDigest,
		IndicatorCount: len(indicators), Status: "VERIFIED",
		PromotionAuthorized: false,
	}
	result.Digest, err = metric.Digest(result)
	return result, err
}
