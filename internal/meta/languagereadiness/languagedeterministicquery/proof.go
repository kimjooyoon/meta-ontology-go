package languagedeterministicquery

func proofs(summary Summary, registryDigest string) []Proof {
	foundationPassed := summary.RegistryDrift == 0 && summary.ConceptBindings == 1 &&
		summary.CodeBindings == 6 && summary.MetricBindings == 18 && summary.UseCaseBindings == 3
	coherencePassed := summary.BindingPlans == 28 && summary.CanonicalReplays == 56 && summary.PermutationReplays == 28
	regressionPassed := summary.LawPlans == 4 && summary.NotSatisfied == 0 && summary.Unresolved == 0 &&
		summary.CandidatePromotions == 0 && summary.UnknownAcceptances == 0 && summary.GraphMutations == 0
	return []Proof{
		{
			Choice: "FOUNDATION", MetaOperation: "bind-versioned-query-plan-registry",
			EvidenceDigest: digestJSON(struct {
				Registry string
				Summary  Summary
			}{registryDigest, summary}), Passed: foundationPassed,
		},
		{
			Choice: "COHERENCE", MetaOperation: "replay-canonical-and-permuted-query-results",
			EvidenceDigest: digestJSON(struct{ Canonical, Permutation int }{summary.CanonicalReplays, summary.PermutationReplays}), Passed: coherencePassed,
		},
		{
			Choice: "REGRESSION", MetaOperation: "reject-unknown-candidate-and-mutation",
			EvidenceDigest: digestJSON(struct{ Laws, Candidates, Unknowns, Mutations int }{summary.LawPlans, summary.CandidatePromotions, summary.UnknownAcceptances, summary.GraphMutations}), Passed: regressionPassed,
		},
	}
}

func allProofsPassed(proofs []Proof) bool {
	if len(proofs) != 3 {
		return false
	}
	for _, proof := range proofs {
		if !proof.Passed {
			return false
		}
	}
	return true
}
