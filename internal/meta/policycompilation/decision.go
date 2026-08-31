package policycompilation

func sameDecision(left, right DecisionResult) bool {
	return left.CaseID == right.CaseID && left.Decision == right.Decision &&
		left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason &&
		left.PolicyDigest == right.PolicyDigest && left.SemanticDigest == right.SemanticDigest &&
		left.Denominator == right.Denominator
}
