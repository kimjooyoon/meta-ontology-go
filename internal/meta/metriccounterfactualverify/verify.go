package metriccounterfactualverify

import (
	"fmt"
	"os"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func Replay(ledger metric.Ledger) (Receipt, error) {
	if err := validateLedger(ledger); err != nil {
		return Receipt{}, err
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
	indicators, err := metric.EvaluateIndicators(ledger.Manifest, ledger.Plan, before, after, receipts, delta)
	if err != nil {
		return Receipt{}, err
	}
	if !artifact.Equal(before, ledger.Before) || !artifact.Equal(after, ledger.After) ||
		!artifact.Equal(receipts, ledger.Receipts) || !artifact.Equal(delta, ledger.Delta) ||
		!artifact.Equal(indicators, ledger.Indicators) || !metric.AllSatisfied(indicators) {
		return Receipt{}, fmt.Errorf("counterfactual replay diverged")
	}
	replayDigest, err := artifact.Digest(struct {
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
		IndicatorCount: len(indicators), Status: "VERIFIED", PromotionAuthorized: false,
	}
	result.Digest, err = artifact.Digest(result)
	return result, err
}
