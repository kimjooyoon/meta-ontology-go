package publicpartialreuse

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Scenario struct {
	Changed        []string
	GraphVariant   string
	ReceiptVariant string
	Option         string
}

func ScenarioForCase(item Case) Scenario {
	changed := []string{item.Changed}
	if item.Changed == "none" {
		changed = nil
	}
	receiptVariant := "valid"
	if item.Option == "tampered" {
		receiptVariant = item.Option
	}
	return Scenario{Changed: changed, GraphVariant: item.GraphVariant, ReceiptVariant: receiptVariant, Option: item.Option}
}

type Metrics struct {
	TestUnitsTotal    int   `json:"test_units_total"`
	TestUnitsExecuted int   `json:"test_units_executed"`
	TestUnitsReused   int   `json:"test_units_reused"`
	BuildExecutions   int   `json:"build_executions"`
	TestExecutions    int   `json:"test_executions"`
	BuildMS           int64 `json:"build_ms"`
	TestMS            int64 `json:"test_ms"`
	WallMS            int64 `json:"wall_ms"`
	PeakRSSKib        int64 `json:"peak_rss_kib"`
}

type Comparisons struct {
	GeneratedBytesEqual    bool `json:"generated_bytes_equal"`
	GeneratedSemanticEqual bool `json:"generated_semantic_equal"`
	TestContractEqual      bool `json:"test_contract_equal"`
	ReceiptBindingEqual    bool `json:"receipt_binding_equal"`
}

type UnknownState struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type CaseReport struct {
	Schema               string        `json:"schema"`
	Operation            string        `json:"operation"`
	CaseID               string        `json:"case_id"`
	Decision             string        `json:"decision"`
	Reason               string        `json:"reason"`
	Unknown              *UnknownState `json:"unknown"`
	ExpectedDecision     string        `json:"expected_decision"`
	Before               Metrics       `json:"before"`
	After                Metrics       `json:"after"`
	ImpactedPartitions   int           `json:"impacted_partitions"`
	UnaffectedPartitions int           `json:"unaffected_partitions"`
	ClosureEdges         int           `json:"closure_edges"`
	ReceiptHits          int           `json:"receipt_hits"`
	ReceiptMisses        int           `json:"receipt_misses"`
	Comparisons          Comparisons   `json:"comparisons"`
	RepositoryWrites     int           `json:"repository_writes"`
	LocalTestExecutions  int           `json:"local_test_executions"`
}

type EvaluationInput struct {
	Policy        Policy
	Case          Case
	Bindings      map[string]Binding
	Receipts      map[string]Receipt
	ReceiptErrors map[string]error
	Execution     Metrics
	Comparisons   Comparisons
}

func Evaluate(input EvaluationInput) (CaseReport, error) {
	policy := input.Policy
	if err := policy.Validate(); err != nil {
		return CaseReport{}, err
	}
	scenario := ScenarioForCase(input.Case)
	declared := declaredEdges(policy, scenario)
	actual := actualEdges(policy)
	closure, closureEdges, closureKnown := impactClosure(policy, scenario.Changed, declared)
	unknowns := make([]UnknownState, 0, 2)
	refuted := make([]string, 0, 2)
	if !closureKnown {
		unknowns = append(unknowns, unknownState("IMPACT", "COMPUTE_CLOSURE", "CHANGED_COMPONENT_IS_UNKNOWN", []string{"canonical_changed_component", "canonical_dependency_graph"}))
	}
	if scenario.GraphVariant == "unbounded" {
		unknowns = append(unknowns, unknownState("IMPACT", "BOUND_CLOSURE", "UNBOUNDED_IMPACT_SCOPE", []string{"bounded_dependency_closure", "finite_partition_ownership"}))
	}
	missing := edgeDifference(actual, declared)
	if len(missing) > 0 {
		if hiddenDependencyOutsideClosure(missing, scenario.Changed, closure) {
			refuted = append(refuted, "CHANGED_DEPENDENCY_OUTSIDE_COMPUTED_CLOSURE")
		} else {
			unknowns = append(unknowns, unknownState("GRAPH", "VALIDATE_DEPENDENCY_EDGES", "MISSING_DEPENDENCY_EDGE", []string{"canonical_dependency_graph", "complete_source_component_relations"}))
		}
	}

	impacted := 0
	for _, partition := range policy.Partitions {
		if closure[partition.ID] {
			impacted++
			continue
		}
		if scenario.ReceiptVariant != "valid" {
			refuted = append(refuted, "TAMPERED_PARTITION_RECEIPT")
			continue
		}
		receipt, ok := input.Receipts[partition.ID]
		if !ok {
			unknowns = append(unknowns, unknownState("REUSE", "LOAD_PARTITION_RECEIPT", "MISSING_OR_STALE_PARTITION_RECEIPT", []string{"immutable_successful_partition_receipt", "explicit_reuse_authorization"}))
			continue
		}
		if err := input.ReceiptErrors[partition.ID]; err != nil {
			refuted = append(refuted, "TAMPERED_PARTITION_RECEIPT")
			continue
		}
		if err := VerifyReceipt(receipt, input.Bindings[partition.ID], partition.ID); err != nil {
			unknowns = append(unknowns, unknownState("REUSE", "VALIDATE_PARTITION_RECEIPT", "STALE_PARTITION_RECEIPT_BINDING", []string{"exact_partition_binding", "same_test_contract", "same_toolchain"}))
		}
	}

	decision := DecisionClosed
	reason := ReuseReason
	var unknown *UnknownState
	if len(refuted) > 0 {
		decision = DecisionRefuted
		reason = RefutedReason
	} else if len(unknowns) > 0 {
		decision = DecisionUnknown
		reason = unknowns[0].Reason
		unknown = &unknowns[0]
	}
	after := input.Execution
	after.TestUnitsTotal = policy.TestUnitCount
	if decision != DecisionClosed {
		after.TestUnitsExecuted, after.TestUnitsReused, after.BuildExecutions, after.TestExecutions = 0, 0, 0, 0
		after.BuildMS, after.TestMS = 0, 0
	} else {
		after.TestUnitsExecuted = impacted
		after.TestUnitsReused = policy.TestUnitCount - impacted
		if after.TestUnitsExecuted != after.TestExecutions {
			return CaseReport{}, errors.New("partial reuse execution metrics do not equal impacted closure")
		}
	}
	return CaseReport{
		Schema: "gooo/public-partial-test-reuse-report/v1", Operation: ReplayOperation, CaseID: input.Case.ID,
		Decision: decision, Reason: reason, Unknown: unknown, ExpectedDecision: input.Case.Decision,
		After: after, ImpactedPartitions: impacted, UnaffectedPartitions: policy.TestUnitCount - impacted, ClosureEdges: closureEdges,
		ReceiptHits: after.TestUnitsReused, ReceiptMisses: policy.TestUnitCount - after.TestUnitsReused, Comparisons: input.Comparisons,
		RepositoryWrites: 0, LocalTestExecutions: 0,
	}, nil
}

func declaredEdges(policy Policy, scenario Scenario) []Edge {
	edges := append([]Edge(nil), policy.Edges...)
	if scenario.GraphVariant != "missing-edge" {
		return edges
	}
	target := strings.TrimPrefix(scenario.Option, "omitted-target=")
	index := -1
	for i, edge := range edges {
		if target != scenario.Option && edge.To == target {
			index = i
			break
		}
	}
	if index < 0 && len(edges) > 0 {
		index = 0
	}
	if index >= 0 {
		edges = append(edges[:index], edges[index+1:]...)
	}
	return edges
}

func impactClosure(policy Policy, changed []string, edges []Edge) (map[string]bool, int, bool) {
	known := map[string]bool{}
	for _, partition := range policy.Partitions {
		known[partition.ID] = true
	}
	for _, component := range policy.Components {
		known[component.ID] = true
	}
	closure := map[string]bool{}
	queue := append([]string(nil), changed...)
	for _, item := range queue {
		if !known[item] {
			return closure, 0, false
		}
		closure[item] = true
	}
	edgesSeen := map[string]bool{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, edge := range edges {
			if edge.From != current || edgesSeen[edge.From+">"+edge.To] {
				continue
			}
			edgesSeen[edge.From+">"+edge.To] = true
			if !closure[edge.To] {
				closure[edge.To] = true
				queue = append(queue, edge.To)
			}
		}
	}
	return closure, len(edgesSeen), true
}

func edgeDifference(actual, declared []Edge) []Edge {
	declaredSet := map[string]bool{}
	for _, edge := range declared {
		declaredSet[edge.From+">"+edge.To] = true
	}
	missing := make([]Edge, 0)
	for _, edge := range actual {
		if !declaredSet[edge.From+">"+edge.To] {
			missing = append(missing, edge)
		}
	}
	return missing
}

func hiddenDependencyOutsideClosure(missing []Edge, changed []string, closure map[string]bool) bool {
	changedSet := map[string]bool{}
	for _, item := range changed {
		changedSet[item] = true
	}
	for _, edge := range missing {
		if changedSet[edge.From] && !closure[edge.To] {
			return true
		}
	}
	return false
}

func unknownState(stage, step, reason string, blocked []string) UnknownState {
	sort.Strings(blocked)
	return UnknownState{Stage: stage, Step: step, Reason: reason, UnknownClass: UnknownClass, NextOperation: UnknownNext, BlockedBy: blocked}
}

func CompareCase(report CaseReport, expected string) error {
	if report.Decision != expected {
		return fmt.Errorf("case %s decision=%s want=%s", report.CaseID, report.Decision, expected)
	}
	if report.Decision == DecisionUnknown {
		if report.Unknown == nil || report.Unknown.Stage == "" || report.Unknown.Step == "" || report.Unknown.Reason == "" || report.Unknown.UnknownClass == "" || report.Unknown.NextOperation == "" || len(report.Unknown.BlockedBy) == 0 {
			return errors.New("UNKNOWN case is missing causal fields")
		}
	} else if report.Unknown != nil {
		return errors.New("non-UNKNOWN case carries UNKNOWN causal state")
	}
	return nil
}
