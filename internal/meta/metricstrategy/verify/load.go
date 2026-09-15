package metricstrategyverify

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	metricverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

type inputSet struct {
	baseline metric.Baseline
	ledger   metric.Ledger
	receipt  metricverify.Receipt
}

func loadInputs(metricsPath, ledgerPath, receiptPath string, plan strategy.Plan) (inputSet, error) {
	baseline, err := metric.LoadBaseline(metricsPath, plan.Repository, plan.SubjectSHA)
	if err != nil {
		return inputSet{}, err
	}
	ledger, err := artifact.ReadJSON[metric.Ledger](ledgerPath)
	if err != nil {
		return inputSet{}, err
	}
	receipt, err := artifact.ReadJSON[metricverify.Receipt](receiptPath)
	if err != nil {
		return inputSet{}, err
	}
	if !metric.ValidLedger(ledger) || !metric.AllSatisfied(ledger.Indicators) || !artifact.Equal(baseline, ledger.Baseline) {
		return inputSet{}, fmt.Errorf("independent strategy input replay diverged")
	}
	if !validInterventionReceipt(receipt) || receipt.Status != "VERIFIED" || receipt.LedgerDigest != ledger.Digest || receipt.SourceMetricsDigest != baseline.SourceMetricsDigest {
		return inputSet{}, fmt.Errorf("independent strategy receipt binding diverged")
	}
	if ledger.RepositoryWorkspaceWrites || ledger.PromotionAuthorized || receipt.RepositoryWorkspaceWrites || receipt.PromotionAuthorized {
		return inputSet{}, fmt.Errorf("independent strategy input boundary is invalid")
	}
	return inputSet{baseline: baseline, ledger: ledger, receipt: receipt}, nil
}
