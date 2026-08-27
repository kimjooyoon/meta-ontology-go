package verify

import (
	"fmt"
	"sort"
)

func finishReconstruction(metrics []cMetric, scenarios []cScenario, recovery sourceReconstruction) ([]cMetric, []cScenario, sourceReconstruction, error) {
	if len(metrics) == 0 || len(scenarios) == 0 {
		return nil, nil, sourceReconstruction{}, fmt.Errorf("consumer recovered no metric/scenario records")
	}
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].id < metrics[j].id })
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].id < scenarios[j].id })
	recovery.Denominator = len(metrics) + len(scenarios)
	recovery.MetricFieldsDenominator = len(metrics) * 8
	recovery.ScenarioDenominator = len(scenarios) * 4
	if recovery.MetricFieldsNumerator != recovery.MetricFieldsDenominator || recovery.ScenarioNumerator != recovery.ScenarioDenominator {
		return nil, nil, sourceReconstruction{}, fmt.Errorf("consumer semantic reconstruction is incomplete")
	}
	return metrics, scenarios, recovery, nil
}
