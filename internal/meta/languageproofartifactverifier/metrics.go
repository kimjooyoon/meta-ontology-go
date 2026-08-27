package languageproofartifactverifier

func MetricIDs() []string {
	return []string{
		"gooo.metric.language.proof-carrying-artifact-cases.v1",
		"gooo.metric.language.proof-carrying-artifact-valid.v1",
		"gooo.metric.language.proof-carrying-artifact-evidence-kinds.v1",
		"gooo.metric.language.proof-carrying-artifact-evidence-links.v1",
		"gooo.metric.language.proof-carrying-artifact-recipes.v1",
		"gooo.metric.language.proof-carrying-artifact-preserved-transitions.v1",
		"gooo.metric.language.proof-carrying-artifact-tampered-rejections.v1",
		"gooo.metric.language.proof-carrying-artifact-missing-evidence.v1",
		"gooo.metric.language.proof-carrying-artifact-byte-only-denials.v1",
		"gooo.metric.language.proof-carrying-artifact-recipe-rejections.v1",
		"gooo.metric.language.proof-carrying-artifact-producer-dependencies.guardrail.v1",
		"gooo.metric.language.proof-carrying-artifact-generated-authority.guardrail.v1",
		"gooo.metric.language.proof-carrying-artifact-semantic-claims.guardrail.v1",
		"gooo.metric.language.proof-carrying-artifact-writes.guardrail.v1",
	}
}

func indicators(summary Summary) []Indicator {
	values := []struct {
		class, proof, operation string
		value, target           int
	}{
		{"OUTCOME", "COHERENCE", "evaluate-proof-carrying-cases", summary.CasesSatisfied, CaseTotal},
		{"OUTCOME", "FOUNDATION", "verify-valid-artifact", summary.ValidArtifacts, 1},
		{"DRIVER", "FOUNDATION", "carry-source-operation-invariant-evidence", summary.EvidenceKindsCarried, EvidenceTotal},
		{"OUTCOME", "COHERENCE", "recheck-evidence-links", summary.ExactEvidenceLinks, EvidenceTotal},
		{"OUTCOME", "FOUNDATION", "match-independent-recipe", summary.RecipeMatches, 1},
		{"DRIVER", "COHERENCE", "preserve-claim-transitions", summary.PreservedTransitions, TransitionTotal},
		{"GUARDRAIL", "REGRESSION", "reject-tampered-evidence", summary.TamperedRejections, 1},
		{"GUARDRAIL", "FOUNDATION", "lower-missing-evidence", summary.MissingEvidenceRejections, 1},
		{"GUARDRAIL", "REGRESSION", "deny-byte-only-authority", summary.ByteOnlyDenials, 1},
		{"GUARDRAIL", "REGRESSION", "reject-recipe-drift", summary.RecipeRejections, 1},
		{"GUARDRAIL", "FOUNDATION", "separate-verifier-from-producer", summary.ProducerDependencies, 0},
		{"GUARDRAIL", "REGRESSION", "deny-generated-authority", summary.GeneratedAuthority, 0},
		{"GUARDRAIL", "FOUNDATION", "bound-semantic-claims", summary.SemanticClaims, 0},
		{"GUARDRAIL", "REGRESSION", "deny-verifier-writes", summary.RepositoryWrites, 0},
	}
	result := make([]Indicator, len(values))
	ids := MetricIDs()
	for index, value := range values {
		result[index] = Indicator{MetricID: ids[index], Class: value.class, ProofChoice: value.proof,
			MetaOperation: value.operation, Value: value.value, Target: value.target, Satisfied: value.value == value.target}
	}
	return result
}

func proofs(summary Summary) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-source-operation-invariant-evidence", EvidenceDigest: digestValue(summary),
			Passed: summary.ValidArtifacts == 1 && summary.EvidenceKindsCarried == EvidenceTotal && summary.RecipeMatches == 1},
		{Choice: "COHERENCE", MetaOperation: "preserve-claim-transitions", EvidenceDigest: digestValue(summary),
			Passed: summary.ExactEvidenceLinks == EvidenceTotal && summary.PreservedTransitions == TransitionTotal},
		{Choice: "REGRESSION", MetaOperation: "deny-byte-only-authority-and-tampering", EvidenceDigest: digestValue(summary),
			Passed: summary.TamperedRejections == 1 && summary.MissingEvidenceRejections == 1 && summary.ByteOnlyDenials == 1 &&
				summary.RecipeRejections == 1 && summary.GeneratedAuthority == 0 && summary.RepositoryWrites == 0},
	}
}
