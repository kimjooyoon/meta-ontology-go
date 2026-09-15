package toolchainconformance

func inspectAll(definitions []SurfaceDefinition, artifacts map[string][]byte,
	expectedHead string) (Summary, []SurfaceResult) {
	summary := Summary{SurfacesTotal: len(definitions)}
	results := make([]SurfaceResult, 0, len(definitions))
	known := make(map[string]bool, len(definitions))
	for _, definition := range definitions {
		known[definition.ID] = true
		summary.CasesTotal += definition.Cases
		summary.IndicatorsTotal += definition.Indicators
		summary.ProofsTotal += definition.Proofs
		raw, ok := artifacts[definition.ID]
		if !ok {
			summary.MissingSurfaces++
			results = append(results, SurfaceResult{
				ID: definition.ID, Schema: definition.Schema, Status: "NOT_SATISFIED",
			})
			continue
		}
		results = append(results, inspectSurface(definition, raw, expectedHead, &summary))
	}
	for id := range artifacts {
		if !known[id] {
			summary.UnexpectedSurfaces++
		}
	}
	summary.ReadinessBPS = basisPoints(summary.SurfacesSatisfied, summary.SurfacesTotal)
	summary.CaseReadinessBPS = basisPoints(summary.CasesSatisfied, summary.CasesTotal)
	summary.ProofReadinessBPS = basisPoints(summary.ProofsPassed, summary.ProofsTotal)
	return summary, results
}

func basisPoints(value, total int) int {
	if total == 0 {
		return 0
	}
	return value * 10000 / total
}
