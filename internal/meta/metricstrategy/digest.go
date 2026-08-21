package metricstrategy

import artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"

func sealPlan(value Plan) (Plan, error) {
	value.Digest = ""
	digest, err := artifact.Digest(value)
	value.Digest = digest
	return value, err
}

func ValidPlanDigest(value Plan) bool {
	digest := value.Digest
	sealed, err := sealPlan(value)
	return err == nil && digest == sealed.Digest
}

func proofChoices() []string {
	return []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
}

