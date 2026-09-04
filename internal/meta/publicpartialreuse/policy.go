package publicpartialreuse

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	PolicySchema        = "gooo/public-partial-test-reuse-policy/v1"
	EvaluatorSchema     = "gooo/public-partial-test-reuse-evaluator/v1"
	PolicyName          = "EXPLICIT_OPT_IN_PARTIAL_TEST_REUSE"
	Operation           = "gooo.test.generated-public-partial-reuse"
	PolicyActivity      = "CreateReceipt"
	ContractVersion     = "v1"
	DecisionClosed      = "CLOSED"
	DecisionUnknown     = "UNKNOWN"
	DecisionRefuted     = "REFUTED"
	ArtifactCount       = 2
	EvidenceCount       = 25
	TestUnitCount       = 2
	DependencyEdgeCount = 2
	CaseCount           = 6
)

type Partition struct {
	ID       string   `json:"id"`
	Activity string   `json:"activity"`
	TestName string   `json:"test_name"`
	Symbols  []string `json:"symbols"`
	Roots    []string `json:"roots"`
}

type Component struct {
	ID   string `json:"id"`
	Root string `json:"root"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Case struct {
	ID             string `json:"id"`
	Decision       string `json:"decision"`
	Changed        string `json:"changed"`
	GraphVariant   string `json:"graph_variant"`
	ReceiptVariant string `json:"receipt_variant"`
	Option         string `json:"option"`
}

type Policy struct {
	Schema              string      `json:"schema"`
	EvaluatorSchema     string      `json:"evaluator_schema"`
	SourceDigest        string      `json:"source_digest"`
	SemanticDigest      string      `json:"semantic_digest"`
	EvaluatorDigest     string      `json:"evaluator_digest"`
	Package             string      `json:"package"`
	Namespace           string      `json:"namespace"`
	Activity            string      `json:"activity"`
	Name                string      `json:"name"`
	Contract            string      `json:"contract"`
	Partitions          []Partition `json:"partitions"`
	Components          []Component `json:"components"`
	Edges               []Edge      `json:"edges"`
	Cases               []Case      `json:"cases"`
	TestUnitCount       int         `json:"test_unit_count"`
	DependencyEdgeCount int         `json:"dependency_edge_count"`
	GeneratedArtifacts  int         `json:"generated_artifacts"`
	EvidenceArtifacts   int         `json:"evidence_artifacts"`
	Journey             []string    `json:"journey"`
	RuntimeRule         string      `json:"runtime_rule"`
	RefutedDominates    bool        `json:"refuted_dominates_unknown"`
	IR                  semantic.IR `json:"-"`
}

type markers map[string][]string

func Load(filename string, source []byte) (Policy, error) {
	if len(source) == 0 {
		return Policy{}, errors.New("partial reuse source is empty")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return Policy{}, errors.New("partial reuse source has syntax diagnostics")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower partial reuse source: %w", err)
	}
	if ir.Package != "partialreuseexample" || ir.Namespace.String() != "partial_reuse_example" {
		return Policy{}, errors.New("partial reuse source is not the canonical example")
	}

	all := markers{}
	partitions := make([]Partition, 0, TestUnitCount)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		current := parseMarkers(node.ValueProgram)
		for key, values := range current {
			all[key] = append(all[key], values...)
		}
		ids := current["partial-reuse-component"]
		if len(ids) == 0 {
			continue
		}
		if len(ids) != 1 || len(current["partial-reuse-test"]) != 1 || len(current["partial-reuse-symbols"]) != 1 || len(current["partial-reuse-roots"]) != 1 {
			return Policy{}, errors.New("partial reuse partition metadata is incomplete")
		}
		partitions = append(partitions, Partition{
			ID: ids[0], Activity: node.Name, TestName: current["partial-reuse-test"][0],
			Symbols: splitCSV(current["partial-reuse-symbols"][0]), Roots: splitCSV(current["partial-reuse-roots"][0]),
		})
	}
	components := make([]Component, 0, len(all["partial-reuse-component-root"]))
	for _, value := range all["partial-reuse-component-root"] {
		fields := strings.Split(value, "|")
		if len(fields) != 2 {
			return Policy{}, fmt.Errorf("partial reuse component root %q is malformed", value)
		}
		components = append(components, Component{ID: strings.TrimSpace(fields[0]), Root: strings.TrimSpace(fields[1])})
	}
	edges := make([]Edge, 0, len(all["partial-reuse-edge"]))
	for _, value := range all["partial-reuse-edge"] {
		fields := strings.Split(value, ">")
		if len(fields) != 2 {
			return Policy{}, fmt.Errorf("partial reuse dependency edge %q is malformed", value)
		}
		edges = append(edges, Edge{From: strings.TrimSpace(fields[0]), To: strings.TrimSpace(fields[1])})
	}
	cases := make([]Case, 0, len(all["partial-reuse-case"]))
	for _, value := range all["partial-reuse-case"] {
		fields := strings.Split(value, "|")
		if len(fields) != 5 {
			return Policy{}, fmt.Errorf("partial reuse case %q is malformed", value)
		}
		cases = append(cases, Case{ID: fields[0], Decision: fields[1], Changed: fields[2], GraphVariant: fields[3], Option: fields[4]})
	}
	policy := Policy{
		Schema: PolicySchema, EvaluatorSchema: EvaluatorSchema, SourceDigest: cache.HashBytes(source).String(),
		SemanticDigest: ir.StableHash(), EvaluatorDigest: evaluatorDigest(), Package: ir.Package, Namespace: ir.Namespace.String(),
		Activity: PolicyActivity, Name: first(all, "partial-reuse-policy"), Contract: first(all, "partial-reuse-schema"),
		Partitions: partitions, Components: components, Edges: edges, Cases: cases,
		TestUnitCount: parseInt(first(all, "partial-reuse-test-unit-count")), DependencyEdgeCount: parseInt(first(all, "partial-reuse-dependency-edge-count")),
		GeneratedArtifacts: parseInt(first(all, "partial-reuse-generated-artifact-count")), EvidenceArtifacts: parseInt(first(all, "partial-reuse-evidence-artifact-count")),
		Journey: strings.Split(first(all, "partial-reuse-journey"), ">"), RuntimeRule: first(all, "partial-reuse-runtime-rule"),
		RefutedDominates: first(all, "partial-reuse-refuted-dominates-unknown") == "true", IR: ir,
	}
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (policy Policy) Validate() error {
	if policy.Schema != PolicySchema || policy.EvaluatorSchema != EvaluatorSchema || policy.Name != PolicyName ||
		policy.Contract != ContractVersion || policy.Activity != PolicyActivity || policy.Package != "partialreuseexample" ||
		policy.Namespace != "partial_reuse_example" || policy.TestUnitCount != TestUnitCount ||
		policy.DependencyEdgeCount != DependencyEdgeCount || policy.GeneratedArtifacts != ArtifactCount ||
		policy.EvidenceArtifacts != EvidenceCount || len(policy.Partitions) != TestUnitCount ||
		len(policy.Components) != 1 || len(policy.Edges) != DependencyEdgeCount || len(policy.Cases) != CaseCount ||
		len(policy.Journey) < 5 || policy.RuntimeRule == "" || !policy.RefutedDominates {
		return errors.New("partial reuse policy identity or denominator is invalid")
	}
	if policy.SourceDigest == "" || policy.SemanticDigest == "" || policy.EvaluatorDigest == "" {
		return errors.New("partial reuse policy digest identity is missing")
	}
	seen := map[string]bool{}
	for _, partition := range policy.Partitions {
		if partition.ID == "" || partition.Activity == "" || partition.TestName == "" || len(partition.Symbols) == 0 || len(partition.Roots) == 0 || seen[partition.ID] {
			return errors.New("partial reuse partition relation is invalid")
		}
		seen[partition.ID] = true
	}
	if seen["orders"] == false || seen["inventory"] == false || policy.Components[0].ID != "shared-contract" || policy.Components[0].Root != "SharedContract" {
		return errors.New("partial reuse component relation is not canonical")
	}
	seenEdges := map[string]bool{}
	for _, edge := range policy.Edges {
		if edge.From == "" || edge.To == "" || seenEdges[edge.From+">"+edge.To] || edge.From != policy.Components[0].ID || !seen[edge.To] {
			return errors.New("partial reuse dependency relation is invalid")
		}
		seenEdges[edge.From+">"+edge.To] = true
	}
	counts := map[string]int{}
	for _, item := range policy.Cases {
		if item.ID == "" || (item.Decision != DecisionClosed && item.Decision != DecisionUnknown && item.Decision != DecisionRefuted) || item.Changed == "" || item.GraphVariant == "" || item.Option == "" {
			return errors.New("partial reuse acceptance case is invalid")
		}
		counts[item.Decision]++
	}
	if counts[DecisionClosed] != 2 || counts[DecisionUnknown] != 2 || counts[DecisionRefuted] != 2 {
		return errors.New("partial reuse acceptance cases are not 2/2/2")
	}
	return nil
}

func (policy Policy) Case(id string) (Case, bool) {
	for _, item := range policy.Cases {
		if item.ID == id {
			return item, true
		}
	}
	return Case{}, false
}

func (policy Policy) CasesCount() int { return len(policy.Cases) }

func (policy Policy) SkeletonDigest() string {
	type skeleton struct {
		Schema           string      `json:"schema"`
		Evaluator        string      `json:"evaluator"`
		Name             string      `json:"name"`
		Contract         string      `json:"contract"`
		Components       []Component `json:"components"`
		Edges            []Edge      `json:"edges"`
		Cases            []Case      `json:"cases"`
		Journey          []string    `json:"journey"`
		RuntimeRule      string      `json:"runtime_rule"`
		RefutedDominates bool        `json:"refuted_dominates_unknown"`
	}
	value := skeleton{PolicySchema, EvaluatorSchema, PolicyName, ContractVersion, append([]Component(nil), policy.Components...), append([]Edge(nil), policy.Edges...), append([]Case(nil), policy.Cases...), append([]string(nil), policy.Journey...), policy.RuntimeRule, policy.RefutedDominates}
	return hashJSON(value)
}

func (policy Policy) DependencyGraphDigest() string {
	edges := append([]Edge(nil), policy.Edges...)
	sort.Slice(edges, func(i, j int) bool { return edges[i].From+">"+edges[i].To < edges[j].From+">"+edges[j].To })
	return hashJSON(edges)
}

func parseMarkers(value string) markers {
	result := markers{}
	for _, part := range strings.Split(value, ";") {
		fields := strings.SplitN(part, "=", 2)
		if len(fields) != 2 || strings.TrimSpace(fields[0]) == "" {
			continue
		}
		result[strings.TrimSpace(fields[0])] = append(result[strings.TrimSpace(fields[0])], strings.TrimSpace(fields[1]))
	}
	return result
}

func first(values markers, key string) string {
	if len(values[key]) == 0 {
		return ""
	}
	return values[key][0]
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func parseInt(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func evaluatorDigest() string {
	return cache.HashBytes([]byte(EvaluatorSchema + "\x00impact-closure-precedence-v1")).String()
}
