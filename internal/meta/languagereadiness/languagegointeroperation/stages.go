package languagegointeroperation

func stages(summary Summary) []StageReceipt {
	sealed := summary.InvalidAcceptances == 0 && summary.UnknownAcceptances == 0 &&
		summary.ImportAcceptances == 0 && summary.EffectfulStages == 0
	return []StageReceipt{
		stage(1, "OBSERVE_CONCEPT_ARTIFACT", "FOUNDATION", "observe-explicit-concept-pass", summary.RegistryDrift == 0),
		stage(2, "BIND_FIXED_REGISTRY", "FOUNDATION", "bind-24-versioned-cases", summary.Executed == 24),
		stage(3, "PROJECT_SEMANTIC_IR", "COHERENCE", "execute-existing-go-generator", summary.GeneratorProjections == 8),
		stage(4, "REIFY_GO_AST", "FOUNDATION", "parse-and-format-go-ast", summary.ASTReifications == 32),
		stage(5, "CHECK_GO_TYPES", "COHERENCE", "check-go-1.27-type-boundary", summary.PositiveAccepted == 16),
		stage(6, "NORMALIZE_GO_API", "COHERENCE", "normalize-and-replay-exported-api", summary.CanonicalReplays == 16 && summary.TypeIdentityReplays == 16),
		stage(7, "BIND_GO_1_27", "COHERENCE", "bind-generic-method-and-alias-evidence", summary.Go127Boundaries == 8 && summary.GenericMethods == 5 && summary.AliasNodes == 2),
		stage(8, "SEAL_AUTHORITY", "REGRESSION", "reject-unknown-import-and-effects", sealed && summary.GuardrailRejections == 8),
	}
}

func stage(ordinal int, name, proof, operation string, passed bool) StageReceipt {
	status := "FAIL"
	if passed {
		status = "PASS"
	}
	return StageReceipt{Ordinal: ordinal, Stage: name, ProofChoice: proof,
		MetaOperation: operation, Status: status, Effects: 0}
}

func allStagesPassed(stages []StageReceipt) bool {
	if len(stages) != 8 {
		return false
	}
	for index, receipt := range stages {
		if receipt.Ordinal != index+1 || receipt.Status != "PASS" || receipt.Effects != 0 {
			return false
		}
	}
	return true
}
