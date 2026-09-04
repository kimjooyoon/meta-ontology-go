package publicresolutionrepair

import (
	"errors"
	"fmt"
	"sort"
)

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
	FullTestOutcomeEqual   bool `json:"full_test_outcome_equal"`
	OverlayBindingEqual    bool `json:"overlay_binding_equal"`
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
	Schema                           string        `json:"schema"`
	Operation                        string        `json:"operation"`
	CaseID                           string        `json:"case_id"`
	Decision                         string        `json:"decision"`
	Reason                           string        `json:"reason"`
	ExpectedDecision                 string        `json:"expected_decision"`
	Unknown                          *UnknownState `json:"unknown"`
	ResolutionBefore                 string        `json:"resolution_before"`
	ResolutionAfter                  string        `json:"resolution_after"`
	ProofMode                        string        `json:"proof_mode"`
	OriginalCounterexamplePreserved  bool          `json:"original_counterexample_preserved"`
	OriginalCounterexampleDecision   string        `json:"original_counterexample_decision"`
	OriginalCounterexampleReason     string        `json:"original_counterexample_reason"`
	GraphEdgesBefore                 int           `json:"graph_edges_before"`
	GraphEdgesAfter                  int           `json:"graph_edges_after"`
	ProposedEdges                    int           `json:"proposed_edges"`
	AuthorizedEdges                  int           `json:"authorized_edges"`
	ImpactedClosureBefore            int           `json:"impacted_closure_before"`
	ImpactedClosureAfter             int           `json:"impacted_closure_after"`
	FalseNegativeImpactedTestsBefore int           `json:"false_negative_impacted_tests_before"`
	FalseNegativeImpactedTestsAfter  int           `json:"false_negative_impacted_tests_after"`
	Fallback                         Metrics       `json:"fallback"`
	OverlayReplay                    Metrics       `json:"overlay_replay"`
	UnchangedPartitionSelectivity    Metrics       `json:"unchanged_partition_selectivity"`
	UnchangedSelectivityProven       bool          `json:"unchanged_partition_selectivity_proven"`
	ContinuityEdges                  int           `json:"continuity_edges"`
	Comparisons                      Comparisons   `json:"comparisons"`
	SafetyImprovement                bool          `json:"safety_improvement"`
	RepositoryWrites                 int           `json:"repository_writes"`
	LocalTestExecutions              int           `json:"local_test_executions"`
}

type EvaluationInput struct {
	Policy               Policy
	Case                 Case
	Counterexample       Counterexample
	Proposal             Proposal
	Authorization        AuthorizationArtifact
	Overlay              GraphOverlay
	Fallback             Metrics
	OverlayReplay        Metrics
	UnchangedSelectivity Metrics
	Comparisons          Comparisons
}

func Evaluate(input EvaluationInput) (CaseReport, error) {
	if err := input.Policy.Validate(); err != nil {
		return CaseReport{}, err
	}
	if input.Case.ID == "" {
		return CaseReport{}, errors.New("semantic repair case is empty")
	}
	report := CaseReport{Schema: ReportSchema, Operation: "gooo.test.generated-public-semantic-resolution-repair", CaseID: input.Case.ID, ExpectedDecision: input.Case.Decision, ResolutionBefore: input.Case.ResolutionFrom, ResolutionAfter: input.Case.ResolutionTo, ProofMode: input.Case.ProofMode, OriginalCounterexamplePreserved: input.Counterexample.Valid, OriginalCounterexampleDecision: input.Counterexample.Decision, OriginalCounterexampleReason: input.Counterexample.Reason, GraphEdgesBefore: len(DeclaredEdges(input.Policy, input.Counterexample, nil)), ProposedEdges: 0, AuthorizedEdges: 0, Fallback: input.Fallback, OverlayReplay: input.OverlayReplay, UnchangedPartitionSelectivity: input.UnchangedSelectivity, Comparisons: input.Comparisons, RepositoryWrites: 0, LocalTestExecutions: 0}
	unknowns := make([]UnknownState, 0, 2)
	refuted := make([]string, 0, 2)
	if !input.Counterexample.Valid {
		refuted = append(refuted, "TAMPERED_COUNTEREXAMPLE")
	}
	if input.Case.RepairVariant == "ambiguous" {
		unknowns = append(unknowns, unknownState("REPAIR", "SYNTHESIZE_PROPOSAL", "INSUFFICIENT_OR_AMBIGUOUS_REPAIR_EVIDENCE", []string{"observed_affected_test", "observed_affected_component", "regression_proof"}))
	}
	if input.Case.RepairVariant == "unsupported" {
		unknowns = append(unknowns, unknownState("REPAIR", "CHECK_PROOF_MODE", "UNSUPPORTED_REPAIR_PROOF_MODE", []string{"typed_proof_mode", "repair_eligibility", "regression_evidence"}))
	}
	if input.Case.Authorization == AuthorizationRejected || (input.Authorization.Decision != "" && input.Authorization.Decision != AuthorizationAuthorized) {
		refuted = append(refuted, "UNAUTHORIZED_REPAIR")
	}
	if input.Case.RepairVariant == "canonical" && input.Authorization.Decision == AuthorizationAuthorized {
		if err := ValidateAuthorization(input.Authorization, input.Proposal); err != nil {
			refuted = append(refuted, "TAMPERED_AUTHORIZATION")
		} else if err := ValidateOverlay(input.Overlay, input.Policy, input.Proposal, input.Authorization); err != nil {
			refuted = append(refuted, "TAMPERED_GRAPH_OVERLAY")
		} else {
			report.ProposedEdges = len([]Edge{{From: input.Proposal.From, To: input.Proposal.To}})
			report.AuthorizedEdges = len(input.Overlay.AddedEdges)
			report.GraphEdgesAfter = input.Overlay.OverlayEdgeCount
			report.ContinuityEdges = input.Overlay.ContinuityEdgeCount
		}
	}
	beforeClosure, beforeEdges, beforeKnown := impactClosure(input.Policy, input.Counterexample.ChangedComponent, DeclaredEdges(input.Policy, input.Counterexample, nil))
	report.ImpactedClosureBefore = partitionCount(input.Policy, beforeClosure)
	if beforeKnown == false {
		unknowns = append(unknowns, unknownState("IMPACT", "COMPUTE_CLOSURE", "CHANGED_COMPONENT_IS_UNKNOWN", []string{"canonical_changed_component", "counterexample_graph"}))
	}
	if input.Case.RepairVariant == "canonical" && len(refuted) == 0 {
		afterClosure, afterEdges, afterKnown := impactClosure(input.Policy, input.Counterexample.ChangedComponent, DeclaredEdges(input.Policy, input.Counterexample, &input.Overlay))
		report.ImpactedClosureAfter = partitionCount(input.Policy, afterClosure)
		if !afterKnown {
			unknowns = append(unknowns, unknownState("IMPACT", "REPLAY_CLOSURE", "REPAIRED_CLOSURE_IS_UNKNOWN", []string{"immutable_graph_overlay", "bounded_dependency_closure"}))
		}
		report.FalseNegativeImpactedTestsBefore = input.Policy.TestUnitCount - report.ImpactedClosureBefore
		report.FalseNegativeImpactedTestsAfter = input.Policy.TestUnitCount - report.ImpactedClosureAfter
		if afterEdges != input.Policy.GraphEdgeCountAfter || report.ImpactedClosureAfter != input.Policy.TestUnitCount || report.FalseNegativeImpactedTestsAfter != 0 || !input.Comparisons.FullTestOutcomeEqual || !input.Comparisons.GeneratedBytesEqual || !input.Comparisons.GeneratedSemanticEqual || !input.Comparisons.TestContractEqual || !input.Comparisons.OverlayBindingEqual || input.OverlayReplay.TestUnitsExecuted != input.Policy.OverlayTestUnitsExecuted || input.OverlayReplay.TestUnitsReused != input.Policy.OverlayTestUnitsReused || input.UnchangedSelectivity.TestUnitsExecuted != input.Policy.SelectivityTestUnitsExecuted || input.UnchangedSelectivity.TestUnitsReused != input.Policy.SelectivityTestUnitsReused {
			unknowns = append(unknowns, unknownState("REPLAY", "VERIFY_REPAIRED_CLOSURE", "REPAIRED_CLOSURE_OR_OUTCOME_EVIDENCE_INCOMPLETE", []string{"full_test_outcomes", "complete_affected_partition_closure", "unaffected_partition_selectivity"}))
		}
		report.UnchangedSelectivityProven = input.UnchangedSelectivity.TestUnitsExecuted == 1 && input.UnchangedSelectivity.TestUnitsReused == 1
		report.SafetyImprovement = report.OriginalCounterexamplePreserved && report.OriginalCounterexampleDecision == DecisionRefuted && report.ImpactedClosureBefore < report.ImpactedClosureAfter && report.FalseNegativeImpactedTestsBefore > report.FalseNegativeImpactedTestsAfter && input.Comparisons.FullTestOutcomeEqual
	}
	if input.Case.ResolutionTo == ResolutionFallback && len(refuted) == 0 && len(unknowns) == 0 {
		if input.Fallback.TestUnitsExecuted != input.Policy.FallbackTestUnitsExecuted || input.Fallback.TestUnitsReused != input.Policy.FallbackTestUnitsReused || input.Fallback.TestExecutions != input.Policy.TestUnitCount || input.Fallback.BuildExecutions != 1 {
			unknowns = append(unknowns, unknownState("FALLBACK", "RUN_FULL_PROJECT", "FULL_PROJECT_FALLBACK_EVIDENCE_INCOMPLETE", []string{"complete_test_contract", "successful_fallback_execution"}))
		}
		report.ImpactedClosureAfter = report.ImpactedClosureBefore
		report.FalseNegativeImpactedTestsBefore = input.Policy.TestUnitCount - report.ImpactedClosureBefore
		report.FalseNegativeImpactedTestsAfter = report.FalseNegativeImpactedTestsBefore
	}
	if input.Case.RepairVariant == "none" && input.Case.ResolutionTo == ResolutionFallback && len(refuted) == 0 && report.Fallback.TestUnitsExecuted == input.Policy.FallbackTestUnitsExecuted && report.Fallback.TestUnitsReused == input.Policy.FallbackTestUnitsReused {
		report.ContinuityEdges = beforeEdges
	}
	if len(refuted) > 0 {
		report.Decision = DecisionRefuted
		report.Reason = RefutedReason
		report.Unknown = nil
	} else if len(unknowns) > 0 {
		report.Decision = DecisionUnknown
		report.Reason = unknowns[0].Reason
		report.Unknown = &unknowns[0]
	} else {
		report.Decision = DecisionClosed
		report.Reason = "SAFE_RESOLUTION_WITH_EXPLICIT_FALLBACK_OR_AUTHORIZED_OVERLAY"
	}
	if report.Decision == DecisionClosed && report.CaseID == "" {
		return CaseReport{}, errors.New("semantic repair closed case has no identity")
	}
	return report, nil
}

func CompareCase(report CaseReport, expected string) error {
	if report.Decision != expected {
		return fmt.Errorf("semantic repair case %s decision=%s want=%s", report.CaseID, report.Decision, expected)
	}
	if report.Decision == DecisionUnknown {
		if report.Unknown == nil || report.Unknown.Stage == "" || report.Unknown.Step == "" || report.Unknown.Reason == "" || report.Unknown.UnknownClass == "" || report.Unknown.NextOperation == "" || len(report.Unknown.BlockedBy) == 0 {
			return errors.New("UNKNOWN repair case is missing all causal fields")
		}
	} else if report.Unknown != nil {
		return errors.New("non-UNKNOWN repair case carries UNKNOWN causal state")
	}
	return nil
}

func impactClosure(policy Policy, changed string, edges []Edge) (map[string]bool, int, bool) {
	known := map[string]bool{}
	for _, partition := range policy.CanonicalEdges {
		known[partition.From] = true
		known[partition.To] = true
	}
	if changed == "" || !known[changed] {
		return nil, 0, false
	}
	closure := map[string]bool{changed: true}
	queue := []string{changed}
	seenEdges := map[string]bool{}
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		for _, edge := range edges {
			key := edge.From + ">" + edge.To
			if edge.From != current || seenEdges[key] {
				continue
			}
			seenEdges[key] = true
			if !closure[edge.To] {
				closure[edge.To] = true
				queue = append(queue, edge.To)
			}
		}
	}
	return closure, len(seenEdges), true
}

func partitionCount(policy Policy, closure map[string]bool) int {
	count := 0
	for _, edge := range policy.CanonicalEdges {
		if closure[edge.To] {
			count++
		}
	}
	return count
}

func unknownState(stage, step, reason string, blocked []string) UnknownState {
	sort.Strings(blocked)
	return UnknownState{Stage: stage, Step: step, Reason: reason, UnknownClass: UnknownClass, NextOperation: UnknownNext, BlockedBy: blocked}
}

func partitionFor(policy Policy, id string) (Partition, bool) {
	for _, partition := range policy.Partitions {
		if partition.ID == id {
			return partition, true
		}
	}
	return Partition{}, false
}
