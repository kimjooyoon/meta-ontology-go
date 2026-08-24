package languagedelivery

func summarize(results []ObligationResult, sources []SourceObservation, decoded decodedEvidence) Summary {
	summary := Summary{
		Coordinates:       coordinates(results),
		MetaBindingsTotal: len(results), SourceReceiptsTotal: len(sourceOrder),
		SelfMintedCredits: 0,
	}
	for _, result := range results {
		if result.MetaOperation != "" {
			summary.MetaBindings++
		}
	}
	for _, source := range sources {
		if source.State == "PASS" {
			summary.SourceReceipts++
		}
		summary.Effects.RepositoryWrites += source.RepositoryWrites
		summary.Effects.MutationAuthority = summary.Effects.MutationAuthority || source.MutationAuthority
	}
	summary.ByClass = groupByClass(results)
	summary.ByOwner = groupByOwner(results)
	summary.InternalReadiness = internalReadiness(decoded.Readiness)
	return summary
}

func coordinates(results []ObligationResult) Coordinates {
	value := Coordinates{Total: len(results)}
	for _, result := range results {
		switch result.Status {
		case StatusSatisfied:
			value.Satisfied++
		case StatusNotImplemented:
			value.NotImplemented++
		case StatusNotSatisfied:
			value.NotSatisfied++
		case StatusUnknown:
			value.Unknown++
		}
	}
	if value.Total != 0 {
		value.BasisPoints = value.Satisfied * 10000 / value.Total
	}
	return value
}
