package operationprovenance

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func reconstructSemanticData(ir semantic.IR) ([]metricSpec, []scenarioSpec, SourceReconstruction, error) {
	metrics := make([]metricSpec, 0)
	scenarios := make([]scenarioSpec, 0)
	reconstruction := SourceReconstruction{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		fields, err := computedFields(node.ValueProgram)
		if err != nil {
			return nil, nil, SourceReconstruction{}, fmt.Errorf("activity %s computes record: %w", node.Name, err)
		}
		if node.ValueProgram[:len("metric")] == "metric" {
			metric, err := metricFromFields(fields)
			if err != nil {
				return nil, nil, SourceReconstruction{}, err
			}
			metrics = append(metrics, metric)
			reconstruction.Numerator++
			reconstruction.MetricFieldsNumerator += len(fields)
			continue
		}
		scenario, err := scenarioFromFields(fields)
		if err != nil {
			return nil, nil, SourceReconstruction{}, err
		}
		scenarios = append(scenarios, scenario)
		reconstruction.Numerator++
		reconstruction.ScenarioNumerator += len(fields)
	}
	if len(metrics) == 0 || len(scenarios) == 0 {
		return nil, nil, SourceReconstruction{}, fmt.Errorf("semantic model has no metric and scenario records")
	}
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].ID < metrics[j].ID })
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].ID < scenarios[j].ID })
	reconstruction.Denominator = len(metrics) + len(scenarios)
	reconstruction.MetricFieldsDenominator = len(metrics) * 8
	reconstruction.ScenarioDenominator = len(scenarios) * 4
	if reconstruction.MetricFieldsNumerator != reconstruction.MetricFieldsDenominator || reconstruction.ScenarioNumerator != reconstruction.ScenarioDenominator {
		return nil, nil, SourceReconstruction{}, fmt.Errorf("semantic record reconstruction is incomplete")
	}
	return metrics, scenarios, reconstruction, nil
}
