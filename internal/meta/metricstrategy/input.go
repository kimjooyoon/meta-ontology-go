package metricstrategy

import (
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	metricverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
)

type inputSet struct {
	baseline metric.Baseline
	ledger   metric.Ledger
	receipt  metricverify.Receipt
}

func loadInputs(metricsPath, ledgerPath, receiptPath, repository, subjectSHA string) (inputSet, error) {
	baseline, err := metric.LoadBaseline(metricsPath, repository, subjectSHA)
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
	inputs := inputSet{baseline: baseline, ledger: ledger, receipt: receipt}
	return inputs, validateInputs(inputs, repository, subjectSHA)
}
