package operationprovenance

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func reconstructSemanticData(ir semantic.IR) ([]metricSpec, []scenarioSpec, SourceReconstruction, []Issue, error) {
	metrics := make([]metricSpec, 0)
	scenarios := make([]scenarioSpec, 0)
	reconstruction := SourceReconstruction{}
	issues := make([]Issue, 0)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		fields, err := computedFields(node.ValueProgram)
		if err != nil {
			issues = append(issues, sourceIssue(node.Name, "SHORT_OR_UNKNOWN_COMPUTES", err.Error()))
			continue
		}
		kind := strings.SplitN(node.ValueProgram, "|", 2)[0]
		if kind == "metric" {
			metric, err := metricFromFields(fields)
			if err != nil {
				issues = append(issues, sourceIssue(node.Name, "INVALID_METRIC_RECORD", err.Error()))
				continue
			}
			metrics = append(metrics, metric)
			reconstruction.Numerator++
			reconstruction.MetricFieldsNumerator += len(fields)
			continue
		}
		scenario, err := scenarioFromFields(fields)
		if err != nil {
			issues = append(issues, sourceIssue(node.Name, "INVALID_SCENARIO_RECORD", err.Error()))
			continue
		}
		scenarios = append(scenarios, scenario)
		reconstruction.Numerator++
		reconstruction.ScenarioNumerator += len(fields)
	}
	if len(metrics) == 0 || len(scenarios) == 0 {
		return nil, nil, SourceReconstruction{}, issues, fmt.Errorf("semantic model has no metric and scenario records")
	}
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].ID < metrics[j].ID })
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].ID < scenarios[j].ID })
	reconstruction.Denominator = len(metrics) + len(scenarios)
	reconstruction.MetricFieldsDenominator = len(metrics) * 8
	reconstruction.ScenarioDenominator = len(scenarios) * 4
	if reconstruction.MetricFieldsNumerator != reconstruction.MetricFieldsDenominator || reconstruction.ScenarioNumerator != reconstruction.ScenarioDenominator {
		return nil, nil, SourceReconstruction{}, issues, fmt.Errorf("semantic record reconstruction is incomplete")
	}
	if len(issues) == 0 {
		issues = nil
	}
	return metrics, scenarios, reconstruction, issues, nil
}

func sourceIssue(activity, reason, detail string) Issue {
	return Issue{Stage: "SOURCE", Step: "reconstruct-computes", Reason: reason, Detail: activity + ":" + detail, Cause: "DIRECT_CAUSE"}
}
