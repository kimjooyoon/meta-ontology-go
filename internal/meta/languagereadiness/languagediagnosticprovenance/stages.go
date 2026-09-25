package languagediagnosticprovenance

func stages(summary Summary) []StepReceipt {
	sealed := summary.UnknownAcceptances == 0 &&
		summary.MissingMapAccepts == 0 && summary.AmbiguousAccepts == 0 &&
		summary.InvalidAcceptances == 0 && summary.EffectfulStages == 0
	return []StepReceipt{
		reportStage(1, "OBSERVE_CONCEPT_ARTIFACT", "FOUNDATION",
			"observe-explicit-concept-pass", summary.ConceptDrift == 0),
		reportStage(2, "BIND_FIXED_REGISTRY", "FOUNDATION",
			"bind-18-versioned-cases", summary.Executed == 18),
		reportStage(3, "CAPTURE_PHYSICAL", "FOUNDATION",
			"capture-go-byte-positions", summary.PhysicalPositions == 10),
		reportStage(4, "RESOLVE_LOGICAL", "COHERENCE",
			"resolve-line-directive-positions", summary.LogicalPositions == 10),
		reportStage(5, "BIND_SEMANTIC", "COHERENCE",
			"reverse-generator-source-map", summary.SemanticBindings == 4),
		reportStage(6, "PROJECT_LSP", "COHERENCE",
			"project-normalized-diagnostics", summary.LSPProjections == 10),
		reportStage(7, "REPLAY_CANONICAL", "COHERENCE",
			"replay-diagnostic-traces", summary.CanonicalReplays == 10),
		reportStage(8, "SEAL_UNKNOWN", "REGRESSION",
			"reject-unknown-provenance", sealed && summary.GuardrailRejections == 8),
	}
}

func reportStage(ordinal int, name, proof, operation string, passed bool) StepReceipt {
	status := "FAIL"
	if passed {
		status = "PASS"
	}
	return StepReceipt{
		Ordinal: ordinal, Stage: name, ProofChoice: proof,
		MetaOperation: operation, Status: status, Effects: 0,
	}
}

func allStagesPassed(stages []StepReceipt) bool {
	if len(stages) != 8 {
		return false
	}
	for index, receipt := range stages {
		if receipt.Ordinal != index+1 || receipt.Status != "PASS" ||
			receipt.Effects != 0 {
			return false
		}
	}
	return true
}
