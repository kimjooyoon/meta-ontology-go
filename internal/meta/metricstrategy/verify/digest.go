package metricstrategyverify

import (
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metricverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func validPlan(value strategy.Plan) bool {
	digest := value.Digest
	value.Digest = ""
	expected, err := artifact.Digest(value)
	return err == nil && digest == expected
}

func validInterventionReceipt(value metricverify.Receipt) bool {
	digest := value.Digest
	value.Digest = ""
	expected, err := artifact.Digest(value)
	return err == nil && digest == expected
}

func sealReceipt(value Receipt) (Receipt, error) {
	value.Digest = ""
	digest, err := artifact.Digest(value)
	value.Digest = digest
	return value, err
}

