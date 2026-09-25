package artifactcoverage

import "fmt"

func evaluateIndicators(definitions []Indicator, summary Summary) ([]KPI, error) {
	values := make(map[string]int, 7)
	values["gooo.metric.meta.canonical-operation-artifact.coverage-bps.v1"] = summary.CanonicalCoverageBasisPoints
	values["gooo.metric.meta.exact-head-artifact.coverage-bps.v1"] = summary.ExactHeadCoverageBasisPoints
	values["gooo.metric.meta.digest-bound-artifact.coverage-bps.v1"] = summary.DigestBoundCoverageBasisPoints
	values["gooo.metric.meta.replay-bound-artifact.coverage-bps.v1"] = summary.ReplayBoundCoverageBasisPoints
	values["gooo.metric.meta.uncovered-artifact-operations.guardrail.v1"] = summary.UncoveredOperations
	values["gooo.metric.meta.ambiguous-artifact-bindings.guardrail.v1"] = summary.AmbiguousOperations
	values["gooo.metric.meta.artifact-observer-writes.guardrail.v1"] = summary.RepositoryWrites
	result := make([]KPI, 0, len(definitions))
	for _, definition := range definitions {
		value, exists := values[definition.MetricID]
		if !exists {
			return nil, fmt.Errorf("indicator %q has no runtime value", definition.MetricID)
		}
		satisfied := definition.Relation == RelationGreaterOrEqual && value >= definition.Target ||
			definition.Relation == RelationLessOrEqual && value <= definition.Target
		result = append(result, KPI{Indicator: definition, Value: value, Satisfied: satisfied})
	}
	return result, nil
}
