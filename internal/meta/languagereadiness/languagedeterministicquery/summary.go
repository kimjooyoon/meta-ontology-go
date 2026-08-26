package languagedeterministicquery

func summarize(definitions []Definition, results []CaseResult, drift int) Summary {
	summary := Summary{Total: FixedTotal, Executed: len(results), RegistryDrift: drift}
	for index, result := range results {
		definition := definitions[index]
		if result.Status == StatusSatisfied {
			summary.Satisfied++
		}
		if definition.Kind == CaseBinding {
			summarizeBinding(&summary, definition, result.Evidence)
		} else {
			summarizeLaw(&summary, result.Evidence)
		}
	}
	summary.NotSatisfied = summary.Total - summary.Satisfied - summary.Unresolved
	if summary.Total > 0 {
		summary.ReadinessBPS = summary.Satisfied * 10000 / summary.Total
	}
	return summary
}

func summarizeBinding(summary *Summary, definition Definition, evidence Evidence) {
	summary.BindingPlans++
	if evidence.RequestDigest != "" && evidence.RequestDigest == evidence.ReplayRequest {
		summary.CanonicalReplays++
	}
	if evidence.ResultDigest != "" && evidence.ResultDigest == evidence.ReplayResult {
		summary.CanonicalReplays++
	}
	if evidence.ResultDigest != "" && evidence.ResultDigest == evidence.PermutationResult {
		summary.PermutationReplays++
	}
	switch definition.BindingClass {
	case BindingConcept:
		summary.ConceptBindings++
	case BindingCode:
		summary.CodeBindings++
	case BindingMetric:
		summary.MetricBindings++
	case BindingUseCase:
		summary.UseCaseBindings++
	}
}

func summarizeLaw(summary *Summary, evidence Evidence) {
	summary.LawPlans++
	if evidence.CandidatePromoted {
		summary.CandidatePromotions++
	}
	if evidence.UnknownAccepted {
		summary.UnknownAcceptances++
	}
	if evidence.GraphMutated {
		summary.GraphMutations++
	}
}
