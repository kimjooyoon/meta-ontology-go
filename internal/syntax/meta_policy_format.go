package syntax

import (
	"fmt"
	"strings"
)

func formatPolicy(output *strings.Builder, policy *PolicyDecl) error {
	if policy == nil {
		return fmt.Errorf("nil policy declaration")
	}
	if err := validateIdentifier(policy.Name, "policy name"); err != nil {
		return err
	}
	if policy.ID == "" {
		return fmt.Errorf("policy id is required")
	}
	fmt.Fprintf(output, "policy %s id %s {", policy.Name, quoteString(policy.ID))
	for _, state := range policy.States {
		if state.Name == "" {
			return fmt.Errorf("policy state name is required")
		}
		fmt.Fprintf(output, "\n    state %s", quoteString(state.Name))
	}
	for _, transition := range policy.Transitions {
		if transition.From == "" || transition.To == "" {
			return fmt.Errorf("policy transition endpoints are required")
		}
		fmt.Fprintf(output, "\n    transition %s -> %s", quoteString(transition.From), quoteString(transition.To))
	}
	for _, current := range policy.Cases {
		if current.Name == "" || current.Resolution == nil {
			return fmt.Errorf("policy case %q requires a resolution", current.Name)
		}
		fmt.Fprintf(output, "\n    case %s {", quoteString(current.Name))
		for _, evidence := range current.Evidence {
			if evidence.Name == "" || evidence.Value == "" {
				return fmt.Errorf("policy case %q has incomplete evidence", current.Name)
			}
			fmt.Fprintf(output, "\n        evidence %s %s", quoteString(evidence.Name), quoteString(evidence.Value))
		}
		resolution := current.Resolution
		fmt.Fprintf(output, "\n        resolution {")
		fmt.Fprintf(output, "\n            decision %s\n            stage %s\n            step %s\n            reason %s", quoteString(resolution.Decision), quoteString(resolution.Stage), quoteString(fmt.Sprint(resolution.Step)), quoteString(resolution.Reason))
		if resolution.DecisionStage != "" {
			fmt.Fprintf(output, "\n            decision_stage %s", quoteString(resolution.DecisionStage))
		}
		if resolution.DecisionStep != 0 {
			fmt.Fprintf(output, "\n            decision_step %s", quoteString(fmt.Sprint(resolution.DecisionStep)))
		}
		if resolution.DecisionReason != "" {
			fmt.Fprintf(output, "\n            decision_reason %s", quoteString(resolution.DecisionReason))
		}
		if resolution.UnknownClass != "" {
			fmt.Fprintf(output, "\n            unknown_class %s", quoteString(resolution.UnknownClass))
		}
		if resolution.NextOperation != "" {
			fmt.Fprintf(output, "\n            next_operation %s", quoteString(resolution.NextOperation))
		}
		if len(resolution.BlockedBy) != 0 {
			output.WriteString("\n            blocked_by")
			for _, blocked := range resolution.BlockedBy {
				output.WriteByte(' ')
				output.WriteString(quoteString(blocked))
			}
		}
		if resolution.Role != "" {
			fmt.Fprintf(output, "\n            role %s", quoteString(resolution.Role))
		}
		if resolution.MetaOperation != "" {
			fmt.Fprintf(output, "\n            meta_operation %s", quoteString(resolution.MetaOperation))
		}
		if resolution.ProofChoice != "" {
			fmt.Fprintf(output, "\n            proof_choice %s", quoteString(resolution.ProofChoice))
		}
		if resolution.Claim != "" {
			fmt.Fprintf(output, "\n            claim %s", quoteString(resolution.Claim))
		}
		output.WriteString("\n        }")
		output.WriteString("\n    }")
	}
	output.WriteString("\n}")
	return nil
}
