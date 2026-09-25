package metricstrategy

import metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"

func rootPolicy(value metric.RootPolicy) RootPolicy {
	return RootPolicy{CountsApplicability: value.CountsApplicability, TopologyApplicability: value.TopologyApplicability, TopologyReason: value.TopologyReason, READMERequirement: value.READMERequirement}
}
