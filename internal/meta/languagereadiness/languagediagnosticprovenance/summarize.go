package languagediagnosticprovenance

func summarize(definitions []Definition, results []CaseResult, conceptDrift int) Summary {
	summary := Summary{
		Total: len(definitions), Executed: len(results),
		ConceptDrift: conceptDrift, ToolchainMatches: 1,
	}
	for _, result := range results {
		accumulateResult(&summary, result)
	}
	summary.NotSatisfied = summary.Total - summary.Satisfied
	summary.Unresolved = summary.NotSatisfied
	if summary.Total > 0 {
		summary.ReadinessBPS = summary.Satisfied * 10000 / summary.Total
	}
	if conceptDrift == 0 {
		summary.ConceptBindings = 1
		summary.CodeBindings = len(fixedCodeBindings)
		summary.MetricBindings = len(fixedMetricBindings)
		summary.UseCaseBindings = len(fixedUseCases)
	}
	return summary
}

func accumulateResult(summary *Summary, result CaseResult) {
	if result.Status == StatusSatisfied {
		summary.Satisfied++
	}
	evidence := result.Evidence
	summary.Traced += boolInt(evidence.Traced)
	if result.Definition.Kind == CaseGuardrail && evidence.Rejected &&
		result.Status == StatusSatisfied {
		summary.GuardrailRejections++
	}
	summary.PhysicalPositions += boolInt(evidence.PhysicalBound)
	summary.LogicalPositions += boolInt(evidence.LogicalBound)
	summary.SemanticBindings += boolInt(evidence.SemanticBound)
	summary.LSPProjections += boolInt(evidence.LSPProjected)
	summary.CanonicalReplays += boolInt(evidence.CanonicalReplay)
	summary.OrderedDiagnostics += boolInt(evidence.OrderedDiagnostics)
	summary.LineDirectiveRemaps += boolInt(evidence.LineDirectiveRemap)
	summary.TypeClassifications += boolInt(evidence.TypeClassified)
	summary.ProvenanceSteps += evidence.ProvenanceSteps
	summary.UnknownAcceptances += boolInt(evidence.UnknownAccepted)
	summary.MissingMapAccepts += boolInt(evidence.MissingMapAccepted)
	summary.AmbiguousAccepts += boolInt(evidence.AmbiguousAccepted)
	summary.InvalidAcceptances += boolInt(evidence.InvalidAccepted)
	summary.EffectfulStages += evidence.Effects
}
