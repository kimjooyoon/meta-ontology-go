package policycompilation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func compileFirstClassPolicy(file *syntax.File, policy semantic.Policy) ([]Rule, DecisionReduction, StructureMetrics, error) {
	if file == nil {
		return nil, DecisionReduction{}, StructureMetrics{}, fmt.Errorf("policy syntax tree is nil")
	}
	if len(policy.Cases) != FixedDenominator || len(policy.Transitions) != ReductionRuleCount {
		return nil, DecisionReduction{}, StructureMetrics{}, fmt.Errorf("first-class policy denominator changed: cases=%d transitions=%d", len(policy.Cases), len(policy.Transitions))
	}
	structure := StructureMetrics{
		GrammarNodeKinds:        len(semanticPolicyNodeKinds),
		ASTNodeCounts:           firstClassASTCounts(file),
		IRNodeCounts:            policy.PolicyNodeCounts(),
		TransitionBindings:      len(policy.Transitions),
		EvidenceBindings:        0,
		ResolutionBindings:      len(policy.Cases),
		MarkerOccurrencesBefore: 8,
		MarkerOccurrencesAfter:  0,
		MarkerImprovement:       "UNKNOWN",
	}
	for _, current := range policy.Cases {
		structure.EvidenceBindings += len(current.Evidence)
	}
	if structure.ASTNodeCounts["policy"] != 1 || structure.ASTNodeCounts["state"] != len(policy.States) || structure.ASTNodeCounts["transition"] != len(policy.Transitions) || structure.ASTNodeCounts["case"] != len(policy.Cases) || structure.ASTNodeCounts["evidence"] != structure.EvidenceBindings || structure.ASTNodeCounts["resolution"] != structure.ResolutionBindings {
		return nil, DecisionReduction{}, StructureMetrics{}, errorsForStructure(structure)
	}
	rules := make([]Rule, 0, len(policy.Cases))
	reduction := DecisionReduction{Schema: ReductionSchema, Rules: make([]DecisionRule, 0, len(policy.Cases))}
	for _, current := range policy.Cases {
		resolution := current.Resolution
		if !knownCondition(current.Name) || !knownDecision(resolution.Decision) {
			return nil, DecisionReduction{}, StructureMetrics{}, fmt.Errorf("case %q has an unsupported condition or decision", current.Name)
		}
		if resolution.UnknownClass != "" || resolution.NextOperation != "" || len(resolution.BlockedBy) != 0 {
			if resolution.Decision != DecisionUnknown || resolution.UnknownClass == "" || resolution.NextOperation == "" || len(resolution.BlockedBy) == 0 {
				return nil, DecisionReduction{}, StructureMetrics{}, fmt.Errorf("case %q has invalid UNKNOWN resolution metadata", current.Name)
			}
		} else if resolution.Decision == DecisionUnknown {
			return nil, DecisionReduction{}, StructureMetrics{}, fmt.Errorf("case %q does not preserve UNKNOWN resolution metadata", current.Name)
		}
		if !policyToken.MatchString(resolution.Stage) || !policyToken.MatchString(resolution.Reason) || !policyToken.MatchString(resolution.DecisionStage) || !policyToken.MatchString(resolution.DecisionReason) || !policyToken.MatchString(resolution.Role) || !policyToken.MatchString(resolution.MetaOperation) || !policyToken.MatchString(resolution.ProofChoice) || !policyToken.MatchString(resolution.Claim) || resolution.Step < 1 || resolution.Step > FixedDenominator || resolution.DecisionStep < 1 || resolution.DecisionStep > FixedDenominator {
			return nil, DecisionReduction{}, StructureMetrics{}, fmt.Errorf("case %q has unsafe resolution metadata", current.Name)
		}
		activityID := fmt.Sprintf("%s/case/%s", policy.ID, strings.ToLower(current.Name))
		rules = append(rules, Rule{ActivityID: activityID, ActivityName: current.Name, Role: resolution.Role, MetaOperation: resolution.MetaOperation, ProofChoice: resolution.ProofChoice, Stage: resolution.Stage, Step: resolution.Step, Reason: resolution.Reason, Claim: resolution.Claim})
		reduction.Rules = append(reduction.Rules, DecisionRule{Condition: current.Name, Decision: resolution.Decision, Stage: resolution.DecisionStage, Step: resolution.DecisionStep, Reason: resolution.DecisionReason, UnknownClass: resolution.UnknownClass, NextOperation: resolution.NextOperation, BlockedBy: append([]string(nil), resolution.BlockedBy...)})
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Step < rules[j].Step })
	for index, rule := range rules {
		if rule.Step != index+1 {
			return nil, DecisionReduction{}, StructureMetrics{}, fmt.Errorf("first-class policy step %d is not present exactly once", index+1)
		}
	}
	if err := validateClaimPredicates(rules); err != nil {
		return nil, DecisionReduction{}, StructureMetrics{}, err
	}
	return rules, reduction, structure, nil
}

var semanticPolicyNodeKinds = []string{"policy", "state", "transition", "case", "evidence", "resolution"}

func firstClassASTCounts(file *syntax.File) map[string]int {
	counts := map[string]int{}
	for _, declaration := range file.Declarations {
		policy, ok := declaration.(*syntax.PolicyDecl)
		if !ok {
			continue
		}
		counts["policy"]++
		counts["state"] += len(policy.States)
		counts["transition"] += len(policy.Transitions)
		counts["case"] += len(policy.Cases)
		for _, current := range policy.Cases {
			counts["evidence"] += len(current.Evidence)
			if current.Resolution != nil {
				counts["resolution"]++
			}
		}
	}
	for _, kind := range semanticPolicyNodeKinds {
		if _, exists := counts[kind]; !exists {
			counts[kind] = 0
		}
	}
	return counts
}

func errorsForStructure(structure StructureMetrics) error {
	return fmt.Errorf("first-class AST/IR node counts disagree: ast=%v ir=%v", structure.ASTNodeCounts, structure.IRNodeCounts)
}
