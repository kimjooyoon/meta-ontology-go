// Package verify is the independent consumer of the provenance receipt. It
// intentionally does not import the producer package: the raw source is
// parsed, lowered, reconstructed, and judged again here.
package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	receiptSchema = "gooo/meta-operation-provenance-receipt/v2"
	reportSchema  = "gooo/meta-operation-provenance-verification/v2"
	toolchain     = "go1.27.0"
)

type lineage struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	EvidencePath  string `json:"evidence_path"`
}

type provenance struct {
	SourceDigest     string `json:"source_digest"`
	SemanticDigest   string `json:"semantic_digest"`
	Producer         string `json:"producer"`
	Consumer         string `json:"consumer"`
	MetaOperation    string `json:"meta_operation"`
	EvidencePath     string `json:"evidence_path"`
	ScenarioMutation string `json:"scenario_mutation"`
}

type issue struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Cause     string   `json:"cause"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}

type claimTransition struct {
	PriorClaim          string     `json:"prior_claim"`
	NextClaim           string     `json:"next_claim"`
	ConformanceDecision string     `json:"conformance_decision"`
	SubjectResolution   string     `json:"subject_resolution"`
	Transition          string     `json:"transition"`
	Stage               string     `json:"stage"`
	Step                string     `json:"step"`
	Reason              string     `json:"reason"`
	EvidenceDigest      string     `json:"evidence_digest"`
	Provenance          provenance `json:"provenance"`
}

type metricResult struct {
	ID                string          `json:"id"`
	Family            string          `json:"family"`
	Claim             string          `json:"claim"`
	Numerator         int             `json:"numerator"`
	Denominator       int             `json:"denominator"`
	Decision          string          `json:"decision"`
	SubjectResolution string          `json:"subject_resolution"`
	EvaluationState   string          `json:"evaluation_state"`
	Lineage           lineage         `json:"lineage"`
	Issue             *issue          `json:"issue,omitempty"`
	Transition        claimTransition `json:"claim_transition"`
}

type graphSummary struct {
	Nodes     int            `json:"nodes"`
	Edges     int            `json:"edges"`
	EdgeKinds map[string]int `json:"edge_kinds"`
}

type scenarioResult struct {
	ID                  string         `json:"id"`
	Mutation            string         `json:"mutation"`
	Graph               graphSummary   `json:"graph"`
	Numerator           int            `json:"numerator"`
	Denominator         int            `json:"denominator"`
	ConformanceDecision string         `json:"conformance_decision"`
	SubjectResolution   string         `json:"subject_resolution"`
	Decisions           map[string]int `json:"decisions"`
	Transitions         map[string]int `json:"transitions"`
	Metrics             []metricResult `json:"metrics"`
}

type sourceReconstruction struct {
	Numerator               int `json:"numerator"`
	Denominator             int `json:"denominator"`
	MetricFieldsNumerator   int `json:"metric_fields_numerator"`
	MetricFieldsDenominator int `json:"metric_fields_denominator"`
	ScenarioNumerator       int `json:"scenario_numerator"`
	ScenarioDenominator     int `json:"scenario_denominator"`
}

type workspaceObservation struct {
	BeforeDigest              string   `json:"before_digest"`
	AfterDigest               string   `json:"after_digest"`
	ChangedPaths              []string `json:"changed_paths,omitempty"`
	RepositoryWorkspaceWrites bool     `json:"repository_workspace_writes"`
	MutationAuthority         bool     `json:"mutation_authority"`
}

type receipt struct {
	Schema                  string               `json:"schema"`
	Toolchain               string               `json:"toolchain"`
	SourceDigest            string               `json:"source_digest"`
	CanonicalSemanticDigest string               `json:"canonical_semantic_digest"`
	SourceReconstruction    sourceReconstruction `json:"source_reconstruction"`
	WorkspaceObservation    workspaceObservation `json:"workspace_observation"`
	Scenarios               []scenarioResult     `json:"scenarios"`
	Digest                  string               `json:"digest"`
}

type cMetric struct {
	id, family, claim, producer, consumer, operation, evidence string
	dependsOn                                                  []string
}

type cScenario struct {
	id, removeRelation, dependency, reason string
}

type cRelation struct{ kind, from, to string }
type cEdge struct{ from, to, kind string }
type cFixture struct {
	id, mutation string
	nodes        map[string]string
	edges        []cEdge
	metrics      []cMetric
}

// Verify takes a receipt, the raw .gooo source, and the consumer source used
// for the import check. A missing raw source is an explicit lower-resolution
// result rather than an assertion that the subject is conforming.
func Verify(payload, source, consumerSource []byte) (map[string]any, error) {
	importCheck := producerImportCheck(consumerSource)
	if len(source) == 0 {
		return unknownReport(importCheck), nil
	}
	var actual receipt
	if err := json.Unmarshal(payload, &actual); err != nil {
		return nil, fmt.Errorf("decode receipt: %w", err)
	}
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return nil, fmt.Errorf("consumer source does not parse")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return nil, fmt.Errorf("consumer lowering failed: %w", err)
	}
	if actual.Schema != receiptSchema || actual.Toolchain != toolchain {
		return nil, fmt.Errorf("receipt schema or toolchain is invalid")
	}
	if actual.SourceDigest != digest(source) || actual.CanonicalSemanticDigest != "sha256:"+ir.StableHash() {
		return nil, fmt.Errorf("raw or canonical semantic source digest is not bound")
	}
	if actual.WorkspaceObservation.BeforeDigest != actual.WorkspaceObservation.AfterDigest || len(actual.WorkspaceObservation.ChangedPaths) != 0 || actual.WorkspaceObservation.RepositoryWorkspaceWrites || actual.WorkspaceObservation.MutationAuthority {
		return nil, fmt.Errorf("repository write observation is not clean")
	}
	metrics, scenarios, reconstruction, err := reconstruct(ir)
	if err != nil {
		return nil, err
	}
	expected := receipt{
		Schema: actual.Schema, Toolchain: actual.Toolchain, SourceDigest: actual.SourceDigest,
		CanonicalSemanticDigest: actual.CanonicalSemanticDigest, SourceReconstruction: reconstruction,
		WorkspaceObservation: actual.WorkspaceObservation,
		Scenarios:            make([]scenarioResult, 0, len(scenarios)),
	}
	for _, scenario := range scenarios {
		expected.Scenarios = append(expected.Scenarios, evaluate(scenario, metrics, actual.SourceDigest, actual.CanonicalSemanticDigest))
	}
	if !reflect.DeepEqual(actual.Scenarios, expected.Scenarios) || actual.SourceReconstruction != expected.SourceReconstruction {
		return nil, fmt.Errorf("receipt differs from independent semantic reconstruction")
	}
	withoutDigest := actual
	withoutDigest.Digest = ""
	wantDigest, err := digestJSON(withoutDigest)
	if err != nil || actual.Digest != wantDigest {
		return nil, fmt.Errorf("receipt digest is not bound")
	}
	result := map[string]any{
		"schema": reportSchema, "status": "VERIFIED", "conformance_decision": "VERIFIED", "subject_resolution": "EXACT",
		"source_digest": actual.SourceDigest, "canonical_semantic_digest": actual.CanonicalSemanticDigest,
		"receipt_digest": actual.Digest, "scenario_count": len(actual.Scenarios), "metric_count": len(metrics) * len(actual.Scenarios),
		"fail_closed_count": 0, "direct_unknowns": 0, "dependency_blocks": 0,
		"transition_counts": map[string]int{}, "source_reconstruction": actual.SourceReconstruction,
		"producer_import": importCheck,
	}
	transitionCounts := result["transition_counts"].(map[string]int)
	for _, scenario := range actual.Scenarios {
		for decision, count := range scenario.Decisions {
			if decision == "FAIL_CLOSED" {
				result["fail_closed_count"] = result["fail_closed_count"].(int) + count
			}
			if decision == "UNKNOWN" {
				for _, metric := range scenario.Metrics {
					if metric.Decision != "UNKNOWN" || metric.Issue == nil {
						continue
					}
					if metric.Issue.Cause == "DEPENDENCY_BLOCK" {
						result["dependency_blocks"] = result["dependency_blocks"].(int) + 1
					} else {
						result["direct_unknowns"] = result["direct_unknowns"].(int) + 1
					}
				}
			}
		}
		for transition, count := range scenario.Transitions {
			transitionCounts[transition] += count
		}
	}
	result["digest"], err = digestJSON(result)
	return result, err
}

func producerImportCheck(source []byte) map[string]any {
	if len(source) == 0 {
		return map[string]any{"numerator": 0, "denominator": 1, "status": "UNKNOWN"}
	}
	if strings.Contains(string(source), "internal/meta/operationprovenance\"") {
		return map[string]any{"numerator": 0, "denominator": 1, "status": "FAIL"}
	}
	return map[string]any{"numerator": 1, "denominator": 1, "status": "PASS"}
}

func unknownReport(importCheck map[string]any) map[string]any {
	result := map[string]any{
		"schema": reportSchema, "status": "UNKNOWN", "conformance_decision": "UNKNOWN", "subject_resolution": "LOWER_RESOLUTION",
		"scenario_count": 0, "metric_count": 0, "fail_closed_count": 0, "direct_unknowns": 0, "dependency_blocks": 0,
		"transition_counts": map[string]int{}, "source_reconstruction": sourceReconstruction{}, "producer_import": importCheck,
		"issue": map[string]any{"stage": "CONSUMER", "step": "parse-source", "reason": "REQUIRED_RAW_SOURCE_MISSING", "cause": "DIRECT_CAUSE"},
	}
	digestValue, _ := digestJSON(result)
	result["digest"] = digestValue
	return result
}

func lower(source []byte) (semantic.IR, error) {
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return semantic.IR{}, fmt.Errorf("source has syntax errors")
	}
	model, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("lower source: %w", err)
	}
	return model, nil
}

func reconstruct(ir semantic.IR) ([]cMetric, []cScenario, sourceReconstruction, error) {
	metrics := make([]cMetric, 0)
	scenarios := make([]cScenario, 0)
	recovery := sourceReconstruction{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		fields, kind, err := parseComputed(node.ValueProgram)
		if err != nil {
			return nil, nil, sourceReconstruction{}, fmt.Errorf("consumer activity %s: %w", node.Name, err)
		}
		switch kind {
		case "metric":
			metric, err := metricFrom(fields)
			if err != nil {
				return nil, nil, sourceReconstruction{}, err
			}
			metrics = append(metrics, metric)
			recovery.Numerator++
			recovery.MetricFieldsNumerator += len(fields)
		case "scenario":
			scenario, err := scenarioFrom(fields)
			if err != nil {
				return nil, nil, sourceReconstruction{}, err
			}
			scenarios = append(scenarios, scenario)
			recovery.Numerator++
			recovery.ScenarioNumerator += len(fields)
		}
	}
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

func parseComputed(value string) (map[string]string, string, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 2 || (parts[0] != "metric" && parts[0] != "scenario") {
		return nil, "", fmt.Errorf("unsupported computes value")
	}
	fields := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, raw, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return nil, "", fmt.Errorf("malformed computes field %q", part)
		}
		if _, exists := fields[key]; exists {
			return nil, "", fmt.Errorf("duplicate computes field %q", key)
		}
		fields[key] = raw
	}
	return fields, parts[0], nil
}

func metricFrom(fields map[string]string) (cMetric, error) {
	keys := []string{"id", "family", "prior_claim", "producer", "consumer", "meta_operation", "evidence_path", "depends_on"}
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return cMetric{}, fmt.Errorf("consumer metric is missing %s", key)
		}
	}
	metric := cMetric{id: fields["id"], family: fields["family"], claim: fields["prior_claim"], producer: fields["producer"], consumer: fields["consumer"], operation: fields["meta_operation"], evidence: fields["evidence_path"]}
	if fields["depends_on"] != "" {
		metric.dependsOn = strings.Split(fields["depends_on"], ",")
	}
	return metric, nil
}

func scenarioFrom(fields map[string]string) (cScenario, error) {
	for _, key := range []string{"id", "remove_relation", "dependency", "reason"} {
		if _, ok := fields[key]; !ok {
			return cScenario{}, fmt.Errorf("consumer scenario is missing %s", key)
		}
	}
	return cScenario{id: fields["id"], removeRelation: fields["remove_relation"], dependency: fields["dependency"], reason: fields["reason"]}, nil
}

func evaluate(scenario cScenario, metrics []cMetric, sourceDigest, semanticDigest string) scenarioResult {
	fixture := cFixture{id: scenario.id, mutation: mutationDescription(scenario), nodes: map[string]string{}, metrics: append([]cMetric(nil), metrics...)}
	for _, metric := range metrics {
		fixture.nodes["metric:"+metric.id] = "metric"
		for _, value := range []struct{ id, kind string }{{metric.producer, "producer"}, {metric.consumer, "consumer"}, {metric.operation, "meta-operation"}, {metric.evidence, "evidence-path"}} {
			if value.id != "" {
				fixture.nodes[value.id] = value.kind
			}
		}
		for _, link := range links(metric) {
			if link.from != "" && link.to != "" {
				fixture.edges = append(fixture.edges, cEdge{from: link.from, to: link.to, kind: link.kind})
			}
		}
	}
	if scenario.removeRelation != "" {
		parts := strings.SplitN(scenario.removeRelation, ":", 2)
		if len(parts) == 2 {
			fixture.edges = removeEdge(fixture.edges, fixture.metrics, parts[0], parts[1])
		}
	}
	if scenario.dependency != "" {
		parts := strings.SplitN(scenario.dependency, ">", 2)
		if len(parts) == 2 {
			for index := range fixture.metrics {
				if fixture.metrics[index].id == parts[1] {
					fixture.metrics[index].dependsOn = append(fixture.metrics[index].dependsOn, parts[0])
				}
			}
		}
	}

	edgeCounts := make(map[string]int)
	edgeKinds := make(map[string]int)
	for _, edge := range fixture.edges {
		if fixture.nodes[edge.from] != "" && fixture.nodes[edge.to] != "" {
			edgeCounts[edge.from+"\x00"+edge.to+"\x00"+edge.kind]++
			edgeKinds[edge.kind]++
		}
	}
	byID := make(map[string]cMetric, len(fixture.metrics))
	for _, metric := range fixture.metrics {
		byID[metric.id] = metric
	}
	memo := make(map[string]metricResult)
	visiting := make(map[string]bool)
	var eval func(string) metricResult
	eval = func(id string) metricResult {
		if result, ok := memo[id]; ok {
			return result
		}
		metric := byID[id]
		if visiting[id] {
			return resultFor(metric, edgeCounts, "UNKNOWN", &issue{Stage: "DEPENDENCY", Step: "detect-cycle", Reason: "DEPENDENCY_CYCLE", Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{id}}, sourceDigest, semanticDigest, fixture)
		}
		visiting[id] = true
		result := resultFor(metric, edgeCounts, "", nil, sourceDigest, semanticDigest, fixture)
		for _, dependency := range metric.dependsOn {
			upstream, ok := byID[dependency]
			if !ok {
				result = resultFor(metric, edgeCounts, "UNKNOWN", &issue{Stage: "DEPENDENCY", Step: "resolve-upstream", Reason: "UPSTREAM_METRIC_MISSING", Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{dependency}}, sourceDigest, semanticDigest, fixture)
				break
			}
			upstreamResult := eval(upstream.id)
			if upstreamResult.Decision != "PASS" {
				result = resultFor(metric, edgeCounts, "UNKNOWN", &issue{Stage: "DEPENDENCY", Step: "propagate-unknown", Reason: "UPSTREAM_" + upstreamResult.Decision, Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{dependency}}, sourceDigest, semanticDigest, fixture)
				break
			}
		}
		visiting[id] = false
		memo[id] = result
		return result
	}
	results := make([]metricResult, 0, len(fixture.metrics))
	decisions := make(map[string]int)
	transitions := make(map[string]int)
	numerator := 0
	for _, metric := range fixture.metrics {
		result := eval(metric.id)
		results = append(results, result)
		decisions[result.Decision]++
		transitions[result.Transition.Transition]++
		numerator += result.Numerator
	}
	decision := "PASS"
	if decisions["FAIL_CLOSED"] > 0 {
		decision = "FAIL_CLOSED"
	} else if decisions["UNKNOWN"] > 0 {
		decision = "UNKNOWN"
	}
	return scenarioResult{ID: fixture.id, Mutation: fixture.mutation, Graph: graphSummary{Nodes: len(fixture.nodes), Edges: len(fixture.edges), EdgeKinds: edgeKinds}, Numerator: numerator, Denominator: len(results) * 4, ConformanceDecision: decision, SubjectResolution: "EXACT", Decisions: decisions, Transitions: transitions, Metrics: results}
}

func mutationDescription(scenario cScenario) string {
	return "remove_relation=" + scenario.removeRelation + ";dependency=" + scenario.dependency + ";reason=" + scenario.reason
}

func links(metric cMetric) []cRelation {
	return []cRelation{{"PRODUCES", metric.producer, "metric:" + metric.id}, {"CONSUMES", "metric:" + metric.id, metric.consumer}, {"OPERATES", metric.operation, "metric:" + metric.id}, {"EVIDENCED_BY", "metric:" + metric.id, metric.evidence}}
}

func removeEdge(edges []cEdge, metrics []cMetric, kind, metricID string) []cEdge {
	var wanted cRelation
	for _, metric := range metrics {
		if metric.id != metricID {
			continue
		}
		for _, link := range links(metric) {
			if link.kind == kind {
				wanted = link
			}
		}
	}
	filtered := make([]cEdge, 0, len(edges))
	for _, edge := range edges {
		if edge.kind == wanted.kind && edge.from == wanted.from && edge.to == wanted.to {
			continue
		}
		filtered = append(filtered, edge)
	}
	return filtered
}

func resultFor(metric cMetric, edgeCounts map[string]int, forcedDecision string, forcedIssue *issue, sourceDigest, semanticDigest string, fixture cFixture) metricResult {
	result := metricResult{ID: metric.id, Family: metric.family, Claim: metric.claim, Denominator: 4, SubjectResolution: "EXACT", EvaluationState: "EVALUATED"}
	for _, link := range links(metric) {
		if edgeCounts[link.from+"\x00"+link.to+"\x00"+link.kind] != 1 {
			continue
		}
		result.Numerator++
		switch link.kind {
		case "PRODUCES":
			result.Lineage.Producer = link.from
		case "CONSUMES":
			result.Lineage.Consumer = link.to
		case "OPERATES":
			result.Lineage.MetaOperation = link.from
		case "EVIDENCED_BY":
			result.Lineage.EvidencePath = link.to
		}
	}
	if forcedDecision != "" {
		result.Decision, result.Issue = forcedDecision, forcedIssue
	} else if result.Numerator == result.Denominator {
		result.Decision = "PASS"
	} else if result.Lineage.Consumer == "" {
		result.Decision, result.Issue = "FAIL_CLOSED", &issue{Stage: "LINEAGE", Step: "connect-consumer", Reason: "REQUIRED_CONSUMER_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	} else if result.Lineage.Producer == "" {
		result.Decision, result.Issue = "UNKNOWN", &issue{Stage: "LINEAGE", Step: "connect-producer", Reason: "REQUIRED_PRODUCER_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	} else if result.Lineage.MetaOperation == "" {
		result.Decision, result.Issue = "UNKNOWN", &issue{Stage: "LINEAGE", Step: "connect-meta-operation", Reason: "REQUIRED_META_OPERATION_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	} else {
		result.Decision, result.Issue = "UNKNOWN", &issue{Stage: "LINEAGE", Step: "connect-evidence-path", Reason: "REQUIRED_EVIDENCE_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	}
	result.Transition = transitionFor(result, sourceDigest, semanticDigest, fixture)
	return result
}

func transitionFor(result metricResult, sourceDigest, semanticDigest string, fixture cFixture) claimTransition {
	transition := claimTransition{PriorClaim: result.Claim, ConformanceDecision: result.Decision, SubjectResolution: result.SubjectResolution, Stage: "CLAIM", Provenance: provenance{SourceDigest: sourceDigest, SemanticDigest: semanticDigest, Producer: result.Lineage.Producer, Consumer: result.Lineage.Consumer, MetaOperation: result.Lineage.MetaOperation, EvidencePath: result.Lineage.EvidencePath, ScenarioMutation: fixture.mutation}}
	switch result.Decision {
	case "PASS":
		transition.NextClaim, transition.Transition, transition.Step, transition.Reason = "DISCHARGED", "DISCHARGED", "discharge-open-claim", "PASS_DISCHARGES_OPEN_CLAIM"
	case "FAIL_CLOSED":
		transition.NextClaim, transition.Transition, transition.Step, transition.Reason = "REFUTED", "REFUTED", "refute-claim", "FAIL_CLOSED_IS_EXPLICIT_REFUTATION"
	default:
		transition.NextClaim, transition.Transition, transition.Step, transition.Reason = "OPEN", "PRESERVED_OPEN", "preserve-open-claim", "UNKNOWN_PRESERVES_OPEN_CLAIM"
	}
	evidence := struct {
		MetricID   string     `json:"metric_id"`
		Decision   string     `json:"decision"`
		Issue      *issue     `json:"issue,omitempty"`
		Provenance provenance `json:"provenance"`
	}{result.ID, result.Decision, result.Issue, transition.Provenance}
	payload, _ := json.Marshal(evidence)
	transition.EvidenceDigest = digest(payload)
	return transition
}

func digest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digest(payload), nil
}
