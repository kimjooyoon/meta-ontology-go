package sourcepolicy

import "sort"

// Evaluate turns observations into deterministic, actionable indicators.
func Evaluate(policy Policy, observations []Observation) (Report, error) {
	if policy.Schema == "" {
		policy.Schema = Schema
	}
	if err := policy.Validate(); err != nil {
		return Report{}, err
	}
	indicators := make([]Indicator, 0, len(observations))
	for _, observation := range observations {
		indicator, err := evaluateObservation(policy, observation)
		if err != nil {
			return Report{}, err
		}
		indicators = append(indicators, indicator)
	}
	sort.Slice(indicators, func(i, j int) bool {
		if indicators[i].MetricID != indicators[j].MetricID {
			return indicators[i].MetricID < indicators[j].MetricID
		}
		if indicators[i].Subject != indicators[j].Subject {
			return indicators[i].Subject < indicators[j].Subject
		}
		return indicators[i].Value < indicators[j].Value
	})
	return Report{Schema: "gooo/indicator-report/v1", Policy: policy, Indicators: indicators}, nil
}

func evaluateObservation(policy Policy, observation Observation) (Indicator, error) {
	definition, err := definitionFor(policy, observation)
	if err != nil {
		return Indicator{}, err
	}
	satisfied := true
	switch definition.relation {
	case RelationLessOrEqual:
		satisfied = observation.Value <= definition.limit
	case RelationEqual:
		satisfied = observation.Value == definition.limit
	}
	producer := observation.Producer
	if producer == "" {
		producer = "meta-observer"
	}
	return Indicator{MetricID: observation.Dimension, Family: definition.family, Subject: observation.Subject, Value: observation.Value, Limit: definition.limit, Relation: definition.relation, Blocking: definition.blocking, Satisfied: satisfied, Proof: definition.proof, Producer: producer, Consumer: definition.consumer, Operation: definition.operation, Detail: observation.Detail}, nil
}
