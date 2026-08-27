package verify

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func reconstruct(ir semantic.IR) ([]cMetric, []cScenario, sourceReconstruction, error) {
	metrics, scenarios := []cMetric{}, []cScenario{}
	recovery := sourceReconstruction{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		fields, kind, err := parseComputed(node.ValueProgram)
		if err != nil {
			return nil, nil, sourceReconstruction{}, fmt.Errorf("consumer activity %s: %w", node.Name, err)
		}
		if kind == "metric" {
			metric, err := metricFrom(fields)
			if err != nil {
				return nil, nil, sourceReconstruction{}, err
			}
			metrics = append(metrics, metric)
			recovery.Numerator++
			recovery.MetricFieldsNumerator += len(fields)
			continue
		}
		scenario, err := scenarioFrom(fields)
		if err != nil {
			return nil, nil, sourceReconstruction{}, err
		}
		scenarios = append(scenarios, scenario)
		recovery.Numerator++
		recovery.ScenarioNumerator += len(fields)
	}
	return finishReconstruction(metrics, scenarios, recovery)
}
