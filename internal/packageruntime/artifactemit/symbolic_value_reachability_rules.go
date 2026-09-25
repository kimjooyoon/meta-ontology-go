package artifactemit

func findSymbolicValueRule(rules []SymbolicValueContractRule, id string) *SymbolicValueContractRule {
	for index := range rules {
		if rules[index].ID == id {
			return &rules[index]
		}
	}
	return nil
}

func symbolicValueCompleteRuleRecognized(rule *SymbolicValueContractRule) bool {
	return rule != nil && rule.Match.Activity == "NON_EMPTY" && rule.Match.Inputs == "NON_EMPTY" &&
		rule.Decision == "READY" && rule.Resolution == "VALUE_EXACT"
}

func symbolicValueMissingRuleRecognized(rule *SymbolicValueContractRule) bool {
	return rule != nil && rule.Match.Activity == "MISSING_OR_EMPTY" && rule.Match.Inputs == "ANY" &&
		rule.Decision == "FAIL_CLOSED" && rule.Resolution == "LOWER_RESOLUTION"
}

func symbolicValueDefaultRecognized(policy SymbolicValueContractDefault) bool {
	return policy.Decision == "FAIL_CLOSED" && policy.Resolution == "LOWER_RESOLUTION" &&
		policy.Reason == "SYMBOLIC_INVOCATION_VALUE_UNMATCHED"
}

func symbolicValueRuleReachability(rule SymbolicValueContractRule, reachability string, reachable bool, role, reason string) SymbolicValueRuleReachability {
	return SymbolicValueRuleReachability{
		ID: rule.ID, Reachability: reachability, ReachableAfterStructuralGate: reachable,
		Role: role, Reason: reason, ProofChoice: rule.ProofChoice, MetaOperation: rule.MetaOperation,
	}
}

func symbolicValueUnknownDefault(policy SymbolicValueContractDefault) SymbolicValueDefaultReachability {
	return SymbolicValueDefaultReachability{
		Reachability: "UNKNOWN", Role: "UNCLASSIFIED", Reason: "SCHEMA_TO_DEFAULT_RELATION_UNSUPPORTED",
		ProofChoice: policy.ProofChoice, MetaOperation: policy.MetaOperation,
	}
}

func summarizeSymbolicValueReachability(rules []SymbolicValueRuleReachability, policy SymbolicValueDefaultReachability) SymbolicValueReachabilitySummary {
	summary := SymbolicValueReachabilitySummary{PolicyBranches: len(rules) + 1}
	for _, rule := range rules {
		switch rule.Reachability {
		case "REACHABLE":
			summary.ReachableRules++
		case "UNREACHABLE":
			summary.DefenseOnlyRules++
		default:
			summary.UnknownPolicyBranches++
		}
	}
	switch policy.Reachability {
	case "REACHABLE":
		summary.ReachableDefaults++
	case "UNREACHABLE":
		summary.DefenseOnlyDefaults++
	default:
		summary.UnknownPolicyBranches++
	}
	return summary
}
