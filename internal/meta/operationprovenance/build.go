package operationprovenance

import (
	"bytes"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Build(source []byte) (Receipt, error) {
	if !bytes.Equal(source, CanonicalSource()) {
		return Receipt{}, fmt.Errorf("Gooo source is not the canonical meta-operation relation")
	}
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return Receipt{}, fmt.Errorf("Gooo source has syntax errors")
	}
	scenarios := []Fixture{
		baseFixture("complete"),
		removeEdge(baseFixture("disconnected"), "MOP-COHERENCE-001", "CONSUMES"),
		removeEdge(baseFixture("direct-unknown"), "MOP-FOUNDATION-001", "EVIDENCED_BY"),
		dependencyFixture(),
	}
	receipt := Receipt{
		Schema: ReceiptSchema, Toolchain: Toolchain, SourceDigest: digestBytes(source),
		Scenarios: make([]ScenarioResult, 0, len(scenarios)),
	}
	for _, scenario := range scenarios {
		result, err := evaluate(scenario)
		if err != nil {
			return Receipt{}, err
		}
		receipt.Scenarios = append(receipt.Scenarios, result)
	}
	var err error
	receipt.Digest, err = digestValue(receipt)
	return receipt, err
}

func baseFixture(id string) Fixture {
	fixture := Fixture{ID: id}
	for _, definition := range definitions() {
		metricID := "metric:" + definition.ID
		producer := "producer:" + definition.ID
		consumer := "consumer:" + definition.ID
		operation := "operation:" + definition.ID
		evidence := "evidence:" + definition.ID
		fixture.Metrics = append(fixture.Metrics, Metric{ID: definition.ID, Family: definition.Family, Claim: definition.Claim})
		fixture.Nodes = append(fixture.Nodes,
			Node{ID: metricID, Kind: "metric"}, Node{ID: producer, Kind: "producer"},
			Node{ID: consumer, Kind: "consumer"}, Node{ID: operation, Kind: "meta-operation"},
			Node{ID: evidence, Kind: "evidence-path"})
		fixture.Edges = append(fixture.Edges,
			Edge{From: producer, To: metricID, Kind: "PRODUCES"},
			Edge{From: metricID, To: consumer, Kind: "CONSUMES"},
			Edge{From: operation, To: metricID, Kind: "OPERATES"},
			Edge{From: metricID, To: evidence, Kind: "EVIDENCED_BY"})
	}
	return fixture
}

func removeEdge(fixture Fixture, metricID, kind string) Fixture {
	wantFrom, wantTo := "metric:"+metricID, ""
	if kind == "CONSUMES" {
		wantTo = "consumer:" + metricID
	} else if kind == "EVIDENCED_BY" {
		wantTo = "evidence:" + metricID
	}
	edges := fixture.Edges[:0]
	for _, edge := range fixture.Edges {
		if edge.From == wantFrom && edge.To == wantTo && edge.Kind == kind {
			continue
		}
		edges = append(edges, edge)
	}
	fixture.Edges = edges
	return fixture
}

func dependencyFixture() Fixture {
	fixture := removeEdge(baseFixture("dependency-blocked"), "MOP-FOUNDATION-001", "EVIDENCED_BY")
	for index := range fixture.Metrics {
		if fixture.Metrics[index].ID == "MOP-REGRESSION-001" {
			fixture.Metrics[index].DependsOn = []string{"MOP-FOUNDATION-001"}
		}
	}
	return fixture
}

type relation struct{ kind, from, to string }

func relations(metricID string) []relation {
	return []relation{
		{kind: "PRODUCES", from: "producer:" + metricID, to: "metric:" + metricID},
		{kind: "CONSUMES", from: "metric:" + metricID, to: "consumer:" + metricID},
		{kind: "OPERATES", from: "operation:" + metricID, to: "metric:" + metricID},
		{kind: "EVIDENCED_BY", from: "metric:" + metricID, to: "evidence:" + metricID},
	}
}

func evaluate(fixture Fixture) (ScenarioResult, error) {
	nodes := make(map[string]bool, len(fixture.Nodes))
	for _, node := range fixture.Nodes {
		nodes[node.ID] = true
	}
	edgeCounts := make(map[string]int)
	for _, edge := range fixture.Edges {
		if nodes[edge.From] && nodes[edge.To] {
			edgeCounts[edge.From+"\x00"+edge.To+"\x00"+edge.Kind]++
		}
	}
	results := make([]MetricResult, 0, len(fixture.Metrics))
	byID := make(map[string]MetricResult)
	for _, metric := range fixture.Metrics {
		result := evaluateMetric(metric, edgeCounts, byID)
		results = append(results, result)
		byID[metric.ID] = result
	}
	edgeKinds := make(map[string]int)
	for _, edge := range fixture.Edges {
		edgeKinds[edge.Kind]++
	}
	decisions := make(map[string]int)
	numerator := 0
	for _, result := range results {
		decisions[result.Decision]++
		numerator += result.Numerator
	}
	return ScenarioResult{
		ID: fixture.ID, Graph: GraphSummary{Nodes: len(fixture.Nodes), Edges: len(fixture.Edges), EdgeKinds: edgeKinds},
		Numerator: numerator, Denominator: len(results) * len(relations(metricDefinitions[0].ID)), Decisions: decisions, Metrics: results,
	}, nil
}

func evaluateMetric(metric Metric, edgeCounts map[string]int, previous map[string]MetricResult) MetricResult {
	result := MetricResult{ID: metric.ID, Family: metric.Family, Claim: metric.Claim, Denominator: len(relations(metric.ID))}
	for _, link := range relations(metric.ID) {
		count := edgeCounts[link.from+"\x00"+link.to+"\x00"+link.kind]
		if count == 1 {
			result.Numerator++
		}
		switch link.kind {
		case "PRODUCES":
			if count == 1 {
				result.Lineage.Producer = link.from
			}
		case "CONSUMES":
			if count == 1 {
				result.Lineage.Consumer = link.to
			}
		case "OPERATES":
			if count == 1 {
				result.Lineage.MetaOperation = link.from
			}
		case "EVIDENCED_BY":
			if count == 1 {
				result.Lineage.EvidencePath = link.to
			}
		}
	}
	for _, dependency := range metric.DependsOn {
		if upstream, ok := previous[dependency]; ok && upstream.Decision == "UNKNOWN" {
			result.Decision, result.EvaluationState = "UNKNOWN", "EVALUATED"
			result.Issue = &Issue{Stage: "DEPENDENCY", Step: "propagate-unknown", Reason: "UPSTREAM_UNKNOWN", Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{dependency}}
			return result
		}
	}
	if result.Numerator == result.Denominator {
		result.Decision, result.EvaluationState = "PASS", "EVALUATED"
		return result
	}
	if result.Lineage.Consumer == "" {
		result.Decision, result.EvaluationState = "FAIL_CLOSED", "EVALUATED"
		result.Issue = &Issue{Stage: "LINEAGE", Step: "connect-consumer", Reason: "DISCONNECTED_METRIC", Cause: "DIRECT_CAUSE"}
		return result
	}
	result.Decision, result.EvaluationState = "UNKNOWN", "EVALUATED"
	reason := "REQUIRED_LINEAGE_LINK_MISSING"
	step := "bind-lineage"
	if result.Lineage.EvidencePath == "" {
		reason, step = "REQUIRED_EVIDENCE_MISSING", "evidence-path"
	}
	result.Issue = &Issue{Stage: "BINDING", Step: step, Reason: reason, Cause: "DIRECT_CAUSE"}
	return result
}
