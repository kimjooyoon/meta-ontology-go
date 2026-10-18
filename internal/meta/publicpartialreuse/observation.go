package publicpartialreuse

import "errors"

func observe(input EvaluationInput) (Observation, int, int) {
	policy, scenario := input.Policy, ScenarioForCase(input.Case)
	declared := declaredEdges(policy, scenario)
	closure, edges, known := impactClosure(policy, scenario.Changed, declared)
	missing := edgeDifference(actualEdges(policy), declared)
	value := Observation{ClosureKnown: known, ScopeBounded: scenario.GraphVariant != "unbounded",
		MissingEdge: len(missing) > 0, HiddenChangedEdge: hiddenDependencyOutsideClosure(missing, scenario.Changed, closure),
		Impacted: make([]bool, len(policy.Partitions)), ReceiptStates: make([]string, len(policy.Partitions))}
	impacted := 0
	for i, partition := range policy.Partitions {
		value.Impacted[i] = closure[partition.ID]
		if value.Impacted[i] {
			impacted++
			value.ReceiptStates[i] = "UNREAD_IMPACTED"
			continue
		}
		value.ReceiptStates[i] = receiptState(input, partition.ID)
	}
	return value, impacted, edges
}

func receiptState(input EvaluationInput, partition string) string {
	if err := input.ReceiptErrors[partition]; err != nil {
		if errors.Is(err, ErrUnobservedToolInputs) {
			return "TOOL_UNOBSERVED"
		}
		if errors.Is(err, ErrInvalidReceipt) {
			return "INVALID"
		}
		return "UNAVAILABLE"
	}
	receipt, exists := input.Receipts[partition]
	if !exists {
		return "MISSING"
	}
	if err := VerifyReceipt(receipt, input.Bindings[partition], partition); err != nil {
		if errors.Is(err, ErrUnobservedToolInputs) {
			return "TOOL_UNOBSERVED"
		}
		if errors.Is(err, ErrInvalidReceipt) {
			return "INVALID"
		}
		return "STALE"
	}
	return "VALID"
}
