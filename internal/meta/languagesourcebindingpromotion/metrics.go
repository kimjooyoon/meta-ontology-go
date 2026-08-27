package languagesourcebindingpromotion

func indicators(report Report) []Indicator {
	s := report.Summary
	return []Indicator{
		metric("cases", "OUTCOME", "FOUNDATION", "evaluate-promotion-contract", s.CasesSatisfied, 5),
		metric("promotions", "OUTCOME", "COHERENCE", "promote-source-binding-claim", s.ExactPromotions, 1),
		metric("exact-claims", "DRIVER", "COHERENCE", "traverse-promotion-claim-graph", s.ExactClaims, 3),
		metric("direct-unknowns", "DRIVER", "FOUNDATION", "locate-direct-missing-evidence", s.DirectUnknowns, 3),
		metric("dependency-blocked", "DRIVER", "COHERENCE", "propagate-blocked-claims", s.DependencyBlocked, 4),
		metric("link-refutations", "GUARDRAIL", "REGRESSION", "reject-evidence-link-mismatch", s.LinkRefutations, 1),
		metric("policy-replays", "DRIVER", "COHERENCE", "replay-gooo-promotion-policy", s.PolicyReplays, 1),
		metric("producer-dependencies.guardrail", "GUARDRAIL", "FOUNDATION", "separate-promotion-import-graph", s.ProducerDependencies, 0),
		metric("semantic-claims.guardrail", "GUARDRAIL", "FOUNDATION", "bound-promotion-claim", s.SemanticClaims, 0),
		metric("writes.guardrail", "GUARDRAIL", "REGRESSION", "deny-promotion-repository-writes", report.RepositoryWrites, 0),
	}
}

func metric(id, class, proof, operation string, value, target int) Indicator {
	return Indicator{MetricID: "gooo.metric.language.source-binding-promotion-" + id + ".v1",
		Class: class, ProofChoice: proof, MetaOperation: operation,
		Value: value, Target: target, Satisfied: value == target}
}

func proofs(report Report) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-gooo-policy-and-direct-unknowns", EvidenceDigest: report.PolicySourceDigest,
			Passed: report.Summary.DirectUnknowns == 3 && report.Summary.ProducerDependencies == 0},
		{Choice: "COHERENCE", MetaOperation: "traverse-source-binding-promotion", EvidenceDigest: digestJSON(report.Cases),
			Passed: report.Summary.ExactPromotions == 1 && report.Summary.ExactClaims == 3 && report.Summary.DependencyBlocked == 4 && report.Summary.PolicyReplays == 1},
		{Choice: "REGRESSION", MetaOperation: "refute-mismatched-evidence-link", EvidenceDigest: digestJSON(report.Summary),
			Passed: report.Summary.LinkRefutations == 1 && report.RepositoryWrites == 0 && !report.MutationAuthority},
	}
}
