package main

import "strings"

const (
	claimDependencyRootProgram = "claim.observe:recoverable"
	claimDependencyEdgePrefix  = "claim.edge:"
)

type claimDependencyProgram struct {
	role  string
	kind  string
	label string
}

func parseClaimDependencyProgram(value string) (claimDependencyProgram, *claimTupleFailure) {
	head, label, ok := strings.Cut(value, "|")
	if !ok || !claimDependencyLabelValid(label) {
		return claimDependencyProgram{}, newClaimTupleFailure("CLAIM_DEPENDENCY_LABEL_INVALID", "RESTORE_NONEMPTY_CLAIM_DEPENDENCY_LABEL")
	}
	if head == claimDependencyRootProgram {
		return claimDependencyProgram{role: claimDependencyRootRole, label: label}, nil
	}
	if !strings.HasPrefix(head, claimDependencyEdgePrefix) {
		return claimDependencyProgram{}, newClaimTupleFailure("CLAIM_DEPENDENCY_PROGRAM_PREFIX_INVALID", "RESTORE_CLAIM_DEPENDENCY_PROGRAM_PREFIX")
	}
	kind := strings.TrimPrefix(head, claimDependencyEdgePrefix)
	switch kind {
	case "requires":
		kind = "REQUIRES"
	case "supports":
		kind = "SUPPORTS"
	case "contradicts":
		kind = "CONTRADICTS"
	case "failure-entailment":
		kind = "FAILURE_ENTAILMENT"
	default:
		return claimDependencyProgram{}, newClaimTupleFailure("CLAIM_DEPENDENCY_EDGE_KIND_UNSUPPORTED", "SELECT_REQUIRES_SUPPORTS_CONTRADICTS_OR_FAILURE_ENTAILMENT")
	}
	return claimDependencyProgram{role: claimDependencyEdgeRole, kind: kind, label: label}, nil
}

func claimDependencyLabelValid(label string) bool {
	if label == "" || strings.Contains(label, "|") {
		return false
	}
	for _, char := range label {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '_' && char != '-' && char != '.' {
			return false
		}
	}
	return true
}
