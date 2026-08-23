package languagepackageruntime

func summarize(definitions []Definition, results []CaseResult) Summary {
	summary := Summary{Total: FixedTotal, Executed: len(results), MetricBindings: FixedIndicators}
	for index, result := range results {
		if result.Satisfied { summary.Satisfied++ } else { summary.NotSatisfied++ }
		if result.Kind == "POSITIVE" {
			if result.Satisfied { summary.PositivePaths++ }
			summary.Packages += result.Packages
			summary.Sources += result.Sources
			summary.Imports += result.Imports
			summary.Initializations += result.Initializations
			summary.EntryBindings += result.EntryBindings
			summary.SemanticBindings += result.SemanticBindings
			summary.CanonicalReplays += result.CanonicalReplays
			summary.OrderInvariantReplays += result.OrderInvariantReplays
			summary.EffectfulOperations += result.Effects
			summary.RepositoryWrites += result.RepositoryWrites
		continue
		}
		if result.Satisfied { summary.GuardrailRejections++ }
		classifyGuardrail(&summary, definitions[index], result)
	}
	return summary
}

func classifyGuardrail(summary *Summary, definition Definition, result CaseResult) {
	if result.Observed == "UNKNOWN_FAILURE" { summary.UnknownObservations++ }
	if result.Observed != "ACCEPTED" { return }
	summary.InvalidAcceptances++
	switch definition.Mutation {
	case "DUPLICATE_PACKAGE", "UNKNOWN_IMPORT", "IMPORT_CYCLE": summary.GraphAcceptances++
	case "HEADER_MISMATCH", "PARSE_ERROR": summary.SourceAcceptances++
	case "UNKNOWN_ENTRY_PACKAGE", "UNKNOWN_ENTRY_ACTIVITY": summary.EntryAcceptances++
	}
}
