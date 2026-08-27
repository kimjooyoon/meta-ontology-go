package languageproofartifactverifier

func MetricIDs() []string {
	return []string{
		"gooo.metric.language.proof-carrying-artifact-cases.v2",
		"gooo.metric.language.proof-carrying-artifact-claim-templates.v2",
		"gooo.metric.language.proof-carrying-artifact-claim-instances.v2",
		"gooo.metric.language.proof-carrying-artifact-valid.v2",
		"gooo.metric.language.proof-carrying-artifact-evidence-kinds.v2",
		"gooo.metric.language.proof-carrying-artifact-evidence-links.v2",
		"gooo.metric.language.proof-carrying-artifact-recipes.v2",
		"gooo.metric.language.proof-carrying-artifact-accepted-transitions.v2",
		"gooo.metric.language.proof-carrying-artifact-preserved-transitions.v2",
		"gooo.metric.language.proof-carrying-artifact-case-discharged.v2",
		"gooo.metric.language.proof-carrying-artifact-case-open.v2",
		"gooo.metric.language.proof-carrying-artifact-case-refuted.v2",
		"gooo.metric.language.proof-carrying-artifact-final-ledger-open.v2",
		"gooo.metric.language.proof-carrying-artifact-final-ledger-discharged.v2",
		"gooo.metric.language.proof-carrying-artifact-tampered-rejections.v2",
		"gooo.metric.language.proof-carrying-artifact-coherent-tamper-rejections.v2",
		"gooo.metric.language.proof-carrying-artifact-missing-evidence.v2",
		"gooo.metric.language.proof-carrying-artifact-byte-only-denials.v2",
		"gooo.metric.language.proof-carrying-artifact-recipe-rejections.v2",
		"gooo.metric.language.proof-carrying-artifact-recipe-only-rejections.v2",
		"gooo.metric.language.proof-carrying-artifact-missing-attachment.v2",
		"gooo.metric.language.proof-carrying-artifact-wrong-attachment.v2",
		"gooo.metric.language.proof-carrying-artifact-unrelated-tamper.v2",
		"gooo.metric.language.proof-carrying-artifact-stale-head.v2",
		"gooo.metric.language.proof-carrying-artifact-unauthorized-consumer.v2",
		"gooo.metric.language.proof-carrying-artifact-semantic-interventions.v2",
		"gooo.metric.language.proof-carrying-artifact-nonsemantic-interventions.v2",
		"gooo.metric.language.proof-carrying-artifact-read-only-authority.v2",
		"gooo.metric.language.proof-carrying-artifact-bundle-only.v2",
		"gooo.metric.language.proof-carrying-artifact-consumer-recheck.v2",
		"gooo.metric.language.proof-carrying-artifact-net-repository-state.v2",
		"gooo.metric.language.proof-carrying-artifact-producer-dependencies.guardrail.v2",
		"gooo.metric.language.proof-carrying-artifact-generated-authority.guardrail.v2",
		"gooo.metric.language.proof-carrying-artifact-semantic-claims.guardrail.v2",
		"gooo.metric.language.proof-carrying-artifact-writes.guardrail.v2",
		"gooo.metric.language.proof-carrying-artifact-mutation-authority.guardrail.v2",
		"gooo.metric.language.proof-carrying-artifact-promotion-authority.guardrail.v2",
		"gooo.metric.language.proof-carrying-artifact-semantic-authority.guardrail.v2",
		"gooo.metric.language.proof-carrying-artifact-core-parser-dependencies.guardrail.v2",
	}
}

func indicators(summary Summary) []Indicator {
	values := []struct {
		class, proof, operation string
		value, target           int
	}{
		{"OUTCOME", "COHERENCE", "evaluate-proof-carrying-cases", summary.CasesSatisfied, CaseTotal},
		{"DRIVER", "FOUNDATION", "declare-unique-claim-templates", summary.ClaimTemplates, ClaimTemplateTotal},
		{"DRIVER", "COHERENCE", "evaluate-claim-instances", summary.ClaimInstances, CaseTotal * ClaimTemplateTotal},
		{"OUTCOME", "FOUNDATION", "verify-valid-artifact", summary.ValidArtifacts, 1},
		{"DRIVER", "FOUNDATION", "carry-source-operation-invariant-evidence", summary.EvidenceKindsCarried, EvidenceTotal},
		{"OUTCOME", "COHERENCE", "recheck-evidence-links", summary.ExactEvidenceLinks, EvidenceTotal},
		{"OUTCOME", "FOUNDATION", "match-independent-recipe", summary.RecipeMatches, 1},
		{"DRIVER", "COHERENCE", "accept-digested-claim-transitions", summary.AcceptedTransitions, TransitionTotal},
		{"DRIVER", "COHERENCE", "preserve-transport-transitions", summary.PreservedTransitions, EvidenceTotal + 1},
		{"DRIVER", "COHERENCE", "count-case-discharged-claims", summary.CaseDischargedClaims, 34},
		{"GUARDRAIL", "FOUNDATION", "count-case-open-claims", summary.CaseOpenClaims, 13},
		{"GUARDRAIL", "REGRESSION", "count-case-refuted-claims", summary.CaseRefutedClaims, 13},
		{"DRIVER", "FOUNDATION", "persist-final-ledger-open-claims", summary.FinalLedgerOpenClaims, ClaimTemplateTotal},
		{"DRIVER", "COHERENCE", "persist-final-ledger-discharged-claims", summary.FinalLedgerDischargedClaims, ClaimTemplateTotal},
		{"GUARDRAIL", "REGRESSION", "reject-tampered-evidence", summary.TamperedRejections, 1},
		{"GUARDRAIL", "REGRESSION", "reject-coherent-tamper", summary.CoherentTamperRejections, 1},
		{"GUARDRAIL", "FOUNDATION", "lower-missing-evidence", summary.MissingEvidenceRejections, 1},
		{"GUARDRAIL", "REGRESSION", "deny-byte-only-authority", summary.ByteOnlyDenials, 1},
		{"GUARDRAIL", "REGRESSION", "reject-recipe-drift", summary.RecipeRejections, 1},
		{"GUARDRAIL", "COHERENCE", "reject-recipe-only-drift", summary.RecipeOnlyRejections, 1},
		{"GUARDRAIL", "FOUNDATION", "lower-missing-attachment", summary.MissingAttachmentRejections, 1},
		{"GUARDRAIL", "COHERENCE", "reject-wrong-attachment", summary.WrongAttachmentRejections, 1},
		{"GUARDRAIL", "REGRESSION", "reject-unrelated-evidence-tamper", summary.UnrelatedEvidenceRejections, 1},
		{"GUARDRAIL", "FOUNDATION", "reject-stale-head", summary.StaleHeadRejections, 1},
		{"GUARDRAIL", "REGRESSION", "deny-unauthorized-consumer", summary.UnauthorizedConsumerDenials, 1},
		{"DRIVER", "COHERENCE", "observe-semantic-intervention", summary.SemanticInterventions, 1},
		{"DRIVER", "COHERENCE", "observe-comment-only-intervention", summary.NonsemanticInterventions, 1},
		{"OUTCOME", "COHERENCE", "scope-consumer-authority", summary.ReadOnlyAuthorities, 1},
		{"DRIVER", "COHERENCE", "recheck-bundle-only", summary.BundleOnlyVerification, summary.BundleOnlyVerification},
		{"DRIVER", "COHERENCE", "consume-attested-target", summary.ConsumerRechecks, summary.ConsumerRechecks},
		{"GUARDRAIL", "FOUNDATION", "observe-net-repository-state", summary.NetRepositoryStateUnchanged, 1},
		{"GUARDRAIL", "FOUNDATION", "separate-verifier-from-producer", summary.ProducerDependencies, 0},
		{"GUARDRAIL", "REGRESSION", "deny-generated-authority", summary.GeneratedAuthority, 0},
		{"GUARDRAIL", "FOUNDATION", "bound-semantic-claims", summary.SemanticClaims, 0},
		{"GUARDRAIL", "REGRESSION", "deny-verifier-writes", summary.RepositoryWrites, 0},
		{"GUARDRAIL", "REGRESSION", "deny-mutation-authority", summary.MutationAuthorities, 0},
		{"GUARDRAIL", "REGRESSION", "deny-promotion-authority", summary.PromotionAuthorities, 0},
		{"GUARDRAIL", "REGRESSION", "deny-semantic-authority", summary.SemanticAuthorities, 0},
		{"GUARDRAIL", "FOUNDATION", "record-core-parser-dependencies", summary.CoreParserDependencies, summary.CoreParserDependencies},
	}
	result := make([]Indicator, len(values))
	ids := MetricIDs()
	for index, value := range values {
		result[index] = Indicator{MetricID: ids[index], Class: value.class, ProofChoice: value.proof,
			MetaOperation: value.operation, Value: value.value, Target: value.target, Satisfied: value.value == value.target}
	}
	return result
}

func proofs(report Report, cases []CaseResult) []Proof {
	valid := validCase(cases)
	if valid == nil {
		return []Proof{{Choice: "FOUNDATION", MetaOperation: "bind-source-bytes-and-projection", Passed: false}, {Choice: "COHERENCE", MetaOperation: "bind-operation-and-claim-relations", Passed: false}, {Choice: "REGRESSION", MetaOperation: "execute-negative-and-intervention-suite", Passed: false}}
	}
	evidence := []string{}
	for _, claim := range valid.Claims {
		evidence = append(evidence, claim.EvidenceDigests...)
	}
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-source-bytes-and-projection", TargetDigest: valid.SourceDigest, Dependency: "checkout.source.gooo->source-bytes-bound", EvidenceDigests: append([]string(nil), valid.Claims[0].EvidenceDigests...), ReceiptDigest: valid.OperationDigest, Passed: valid.ObservedDecision == "PASS" && valid.SourceDigest != ""},
		{Choice: "COHERENCE", MetaOperation: "bind-operation-and-claim-relations", TargetDigest: valid.OperationDigest, Dependency: "source-bytes-bound->operation-receipt-bound->recipe-match->consumer-authority", EvidenceDigests: evidence, ReceiptDigest: valid.OperationDigest, Passed: valid.OperationDigest != "" && len(report.Transitions) == TransitionTotal},
		{Choice: "REGRESSION", MetaOperation: "execute-negative-and-intervention-suite", TargetDigest: report.WriteSet.Digest, Dependency: "negative-cases->claim-local-effects->authority-boundary", EvidenceDigests: regressionEvidence(cases, report.Interventions), ReceiptDigest: valid.OperationDigest, Passed: report.Summary.CasesSatisfied == CaseTotal && report.Summary.SemanticInterventions == 1 && report.Summary.NonsemanticInterventions == 1},
	}
}

func regressionEvidence(cases []CaseResult, interventions []InterventionResult) []string {
	result := make([]string, 0, len(cases)+len(interventions))
	for _, item := range cases {
		if item.ID != "valid-proof-carrying-artifact" && item.ArtifactDigest != "" {
			result = append(result, item.ArtifactDigest)
		}
	}
	for _, item := range interventions {
		if item.OperationReceiptDigestAfter != "" {
			result = append(result, item.OperationReceiptDigestAfter)
		}
	}
	return result
}
