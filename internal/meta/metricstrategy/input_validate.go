package metricstrategy

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	metricverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
)

func validateInputs(inputs inputSet, repository, subjectSHA string) error {
	ledger, receipt := inputs.ledger, inputs.receipt
	if ledger.Repository != repository || ledger.SubjectSHA != subjectSHA || !metric.ValidLedger(ledger) {
		return fmt.Errorf("metric strategy intervention subject is invalid")
	}
	if !artifact.Equal(inputs.baseline, ledger.Baseline) || !metric.AllSatisfied(ledger.Indicators) {
		return fmt.Errorf("metric strategy baseline or indicators diverged")
	}
	if ledger.RepositoryWorkspaceWrites || ledger.PromotionAuthorized || !rootPolicyOK(ledger.Baseline.RootPolicy) {
		return fmt.Errorf("metric strategy intervention boundary is invalid")
	}
	if !validReceipt(receipt) || receipt.Schema != metricverify.ReceiptSchema || receipt.Status != "VERIFIED" {
		return fmt.Errorf("metric strategy intervention receipt is invalid")
	}
	if receipt.LedgerDigest != ledger.Digest || receipt.SourceMetricsDigest != ledger.Baseline.SourceMetricsDigest {
		return fmt.Errorf("metric strategy intervention receipt binding diverged")
	}
	if receipt.RepositoryWorkspaceWrites || receipt.PromotionAuthorized || receipt.ProjectionCount != len(ledger.Projections) || receipt.IndicatorCount != len(ledger.Indicators) {
		return fmt.Errorf("metric strategy intervention receipt boundary diverged")
	}
	return nil
}

func validReceipt(value metricverify.Receipt) bool {
	digest := value.Digest
	value.Digest = ""
	expected, err := artifact.Digest(value)
	return err == nil && digest == expected
}

func rootPolicyOK(value metric.RootPolicy) bool {
	return value.CountsApplicability == "OBSERVED" && value.TopologyApplicability == "NOT_APPLICABLE" && value.TopologyReason == "ROOT_TOPOLOGY_EXEMPT" && value.READMERequirement == "NOT_APPLICABLE"
}

