package metricinterventionverify

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	counterverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricintervention"
)

func Replay(metricsPath string, ledger metric.Ledger) (Receipt, error) {
	if !metric.ValidLedger(ledger) || ledger.PromotionAuthorized || ledger.RepositoryWorkspaceWrites {
		return Receipt{}, fmt.Errorf("metric intervention ledger boundary is invalid")
	}
	replayedCounterfactual, err := counterverify.Replay(ledger.Counterfactual)
	if err != nil {
		return Receipt{}, err
	}
	if !artifact.Equal(replayedCounterfactual, ledger.CounterfactualVerification) {
		return Receipt{}, fmt.Errorf("embedded counterfactual replay diverged")
	}
	expected, err := metric.Generate(metricsPath, ledger.Repository, ledger.SubjectSHA)
	if err != nil {
		return Receipt{}, err
	}
	if !artifact.Equal(expected, ledger) || !metric.AllSatisfied(ledger.Indicators) {
		return Receipt{}, fmt.Errorf("metric intervention deterministic replay diverged")
	}
	receipt := Receipt{Schema: ReceiptSchema, LedgerDigest: ledger.Digest, SourceMetricsDigest: ledger.Baseline.SourceMetricsDigest, CounterfactualReplayDigest: replayedCounterfactual.ReplayDigest, ProjectionCount: len(ledger.Projections), IndicatorCount: len(ledger.Indicators), Status: "VERIFIED", RepositoryWorkspaceWrites: false, PromotionAuthorized: false}
	receipt.Digest, err = receiptDigest(receipt)
	return receipt, err
}

func receiptDigest(value Receipt) (string, error) {
	value.Digest = ""
	return artifact.Digest(value)
}
