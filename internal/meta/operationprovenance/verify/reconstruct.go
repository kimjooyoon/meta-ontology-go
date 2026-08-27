package verify

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func reconstruct(ir semantic.IR) ([]cMetric, []cScenario, sourceReconstruction, []issue, error) {
	metrics, scenarios := []cMetric{}, []cScenario{}
	recovery := sourceReconstruction{}
	issues := []issue{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		fields, kind, err := parseComputed(node.ValueProgram)
		if err != nil {
			issues = append(issues, sourceIssue(node.Name, "SHORT_OR_UNKNOWN_COMPUTES", err.Error()))
			continue
		}
		if kind == "metric" {
			metric, err := metricFrom(fields)
			if err != nil {
				issues = append(issues, sourceIssue(node.Name, "INVALID_METRIC_RECORD", err.Error()))
				continue
			}
			metrics = append(metrics, metric)
			recovery.Numerator++
			recovery.MetricFieldsNumerator += len(fields)
			continue
		}
		scenario, err := scenarioFrom(fields)
		if err != nil {
			issues = append(issues, sourceIssue(node.Name, "INVALID_SCENARIO_RECORD", err.Error()))
			continue
		}
		scenarios = append(scenarios, scenario)
		recovery.Numerator++
		recovery.ScenarioNumerator += len(fields)
	}
	return finishReconstruction(metrics, scenarios, recovery, issues)
}

func sourceIssue(activity, reason, detail string) issue {
	return issue{Stage: "SOURCE", Step: "reconstruct-computes", Reason: reason, Detail: activity + ":" + detail, Cause: "DIRECT_CAUSE"}
}
