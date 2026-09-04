package bidir

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func lowerSyntaxPolicies(ctx context.Context, ir *semantic.IR, file *syntax.File) error {
	if file == nil {
		return nil
	}
	for _, declaration := range syntaxDeclarations(file) {
		if err := checkLowerContext(ctx); err != nil {
			return err
		}
		policy, ok := declaration.(*syntax.PolicyDecl)
		if !ok {
			continue
		}
		lowered, err := lowerSyntaxPolicy(policy)
		if err != nil {
			return err
		}
		ir.Policies = append(ir.Policies, lowered)
	}
	return nil
}

func lowerSyntaxPolicy(policy *syntax.PolicyDecl) (semantic.Policy, error) {
	if policy == nil {
		return semantic.Policy{}, fmt.Errorf("nil policy declaration")
	}
	lowered := semantic.Policy{
		ID: semantic.ID(policy.ID), Name: policy.Name, Span: toSemanticSpan(toSourceSpan(policy.Span)),
		States:      make([]semantic.PolicyState, len(policy.States)),
		Transitions: make([]semantic.PolicyTransition, len(policy.Transitions)),
		Cases:       make([]semantic.PolicyCase, len(policy.Cases)),
	}
	for index, state := range policy.States {
		lowered.States[index] = semantic.PolicyState{Name: state.Name, Span: toSemanticSpan(toSourceSpan(state.Span))}
	}
	for index, transition := range policy.Transitions {
		lowered.Transitions[index] = semantic.PolicyTransition{From: transition.From, To: transition.To, Span: toSemanticSpan(toSourceSpan(transition.Span))}
	}
	for index, current := range policy.Cases {
		loweredCase := semantic.PolicyCase{Name: current.Name, Span: toSemanticSpan(toSourceSpan(current.Span)), Evidence: make([]semantic.PolicyEvidence, len(current.Evidence))}
		for evidenceIndex, evidence := range current.Evidence {
			loweredCase.Evidence[evidenceIndex] = semantic.PolicyEvidence{Name: evidence.Name, Value: evidence.Value, Span: toSemanticSpan(toSourceSpan(evidence.Span))}
		}
		if current.Resolution == nil {
			return semantic.Policy{}, fmt.Errorf("policy case %q has no resolution", current.Name)
		}
		resolution := current.Resolution
		loweredCase.Resolution = semantic.PolicyResolution{
			Decision: resolution.Decision, Stage: resolution.Stage, Step: resolution.Step, Reason: resolution.Reason,
			DecisionStage: resolution.DecisionStage, DecisionStep: resolution.DecisionStep, DecisionReason: resolution.DecisionReason,
			UnknownClass: resolution.UnknownClass, NextOperation: resolution.NextOperation,
			BlockedBy: append([]string(nil), resolution.BlockedBy...), Role: resolution.Role,
			MetaOperation: resolution.MetaOperation, ProofChoice: resolution.ProofChoice, Claim: resolution.Claim,
			Span: toSemanticSpan(toSourceSpan(resolution.Span)),
		}
		lowered.Cases[index] = loweredCase
	}
	return lowered.Normalized()
}
