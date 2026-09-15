package languagegointeroperation

func summarize(definitions []Definition, results []CaseResult, drift int) Summary {
	summary := Summary{Total: FixedTotal, Executed: len(results), RegistryDrift: drift, ToolchainMatches: 1}
	if drift == 0 {
		summary.ConceptBindings, summary.CodeBindings = 1, len(fixedCodeBindings)
		summary.MetricBindings, summary.UseCaseBindings = len(fixedMetricBindings), len(fixedUseCases)
	}
	for index, result := range results {
		if result.Status == StatusSatisfied {
			summary.Satisfied++
		}
		summarizeEvidence(&summary, definitions[index], result)
	}
	summary.NotSatisfied = summary.Total - summary.Satisfied - summary.Unresolved
	if summary.Total > 0 {
		summary.ReadinessBPS = summary.Satisfied * 10000 / summary.Total
	}
	return summary
}

func summarizeEvidence(summary *Summary, definition Definition, result CaseResult) {
	evidence := result.Evidence
	if definition.Kind == CaseGuardrail && result.Status == StatusSatisfied && evidence.Rejected {
		summary.GuardrailRejections++
	}
	if definition.Kind != CaseGuardrail && result.Status == StatusSatisfied {
		summarizePositive(summary, definition, evidence)
	}
	summary.InvalidAcceptances += boolInt(evidence.InvalidAccepted)
	summary.UnknownAcceptances += boolInt(evidence.UnknownAccepted)
	summary.ImportAcceptances += boolInt(evidence.ImportAccepted)
	summary.EffectfulStages += evidence.Effects
}

func summarizePositive(summary *Summary, definition Definition, evidence Evidence) {
	summary.PositiveAccepted++
	if definition.Kind == CaseGenerator {
		summary.GeneratorProjections++
		summary.SourceMaps += boolInt(evidence.SourceMapMappings > 0)
	} else {
		summary.Go127Boundaries++
	}
	summary.CanonicalReplays += boolInt(evidence.CanonicalReplay)
	summary.TypeIdentityReplays += boolInt(evidence.TypeIdentityReplay)
	summary.GenericMethods += evidence.GenericMethods
	summary.AliasNodes += evidence.AliasNodes
	summary.ASTReifications += evidence.ASTReifications
}
