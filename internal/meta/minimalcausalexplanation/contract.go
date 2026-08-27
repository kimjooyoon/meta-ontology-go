package minimalcausalexplanation

const (
	requestEvidence   = "evidence.request.accepted"
	policyEvidence    = "evidence.policy.allowed"
	resultEvidence    = "evidence.result.matches"
	noiseEvidence     = "evidence.audit.noise"
	requestPolicyEdge = "edge.request.policy"
	policyResultEdge  = "edge.policy.result"
)

func CanonicalProgram() MetaProgram {
	return MetaProgram{
		Schema:               SourceSchema,
		Producer:             "gooo://meta/minimal-causal-explanation/evaluator",
		Consumer:             "gooo://meta/minimal-causal-explanation/ci-judge",
		IndicatorDenominator: IndicatorTotal,
		MetaOperations: []MetaOperation{
			{ID: "bind-source", Activity: "BindGoooSource", Producer: "source-reader", Consumer: "causal-evaluator", ProofChoice: "FOUNDATION"},
			{ID: "freeze-graph", Activity: "FreezeCausalGraph", Producer: "causal-evaluator", Consumer: "path-checker", ProofChoice: "FOUNDATION"},
			{ID: "evaluate-sufficiency", Activity: "EvaluatePathSufficiency", Producer: "path-checker", Consumer: "counterfactual-checker", ProofChoice: "COHERENCE"},
			{ID: "minimize-path", Activity: "MinimizeByRemoval", Producer: "counterfactual-checker", Consumer: "path-checker", ProofChoice: "COHERENCE"},
			{ID: "judge-receipt", Activity: "JudgeExplanationReceipt", Producer: "independent-judge", Consumer: "ci-minimal-causal-explanation", ProofChoice: "REGRESSION"},
			{ID: "preserve-claims", Activity: "PreserveClaimTransitions", Producer: "claim-ledger", Consumer: "ci-minimal-causal-explanation", ProofChoice: "REGRESSION"},
		},
	}
}

func CanonicalGraph() CausalGraph {
	graph := CausalGraph{
		Schema:       GraphSchema,
		DecisionRule: "PASS iff request.accepted AND policy.allowed AND result.matches are present",
		Nodes: []CausalNode{
			{ID: requestEvidence, Role: "DECISION_INPUT", Producer: "request-observer", Consumer: "causal-evaluator"},
			{ID: policyEvidence, Role: "DECISION_INPUT", Producer: "policy-checker", Consumer: "causal-evaluator"},
			{ID: resultEvidence, Role: "DECISION_INPUT", Producer: "result-observer", Consumer: "causal-evaluator"},
			{ID: noiseEvidence, Role: "NON_CAUSAL_LOG", Producer: "audit-sampler", Consumer: "audit-archive"},
		},
		Edges: []CausalEdge{
			{ID: requestPolicyEdge, From: requestEvidence, To: policyEvidence, Relation: "ENABLES"},
			{ID: policyResultEdge, From: policyEvidence, To: resultEvidence, Relation: "CONSTRAINS"},
		},
	}
	graph.Digest = graphDigest(graph)
	return graph
}

func graphDigest(graph CausalGraph) string {
	graph.Digest = ""
	digest, _ := digestValue(graph)
	return digest
}

func canonicalEvidence() []string {
	return []string{requestEvidence, policyEvidence, resultEvidence}
}

func hasEvidence(evidence []string, wanted string) bool {
	for _, item := range evidence {
		if item == wanted {
			return true
		}
	}
	return false
}

func decisionForEvidence(evidence []string) string {
	for _, wanted := range canonicalEvidence() {
		if !hasEvidence(evidence, wanted) {
			return DecisionFailClosed
		}
	}
	return DecisionPass
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func claimIDs() []string {
	return []string{
		"claim.source-bound",
		"claim.graph-closed",
		"claim.path-sufficient",
		"claim.path-minimal",
		"claim.counterfactual-difference",
		"claim.read-only-preserved",
	}
}
