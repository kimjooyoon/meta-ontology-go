package proposalpromotion

const (
	producer  = "internal/meta/languagereadiness/proposalpromotion"
	consumer  = "internal/meta/languagereadiness"
	operation = "promote-verified-change-proposal"
)

func buildIndicators(summary Summary, source Source) []Indicator {
	return []Indicator{
		indicator("gooo.metric.language.autonomous-change-proposal-promotion-bps.v1",
			"OUTCOME", "COHERENCE", summary.ReadinessBPS, 10_000),
		indicator("gooo.metric.language.autonomous-change-proposal-valid-predecessors.v1",
			"DRIVER", "FOUNDATION", summary.ValidPredecessors, 1),
		indicator("gooo.metric.language.autonomous-change-proposal-contract-bps.v1",
			"DRIVER", "COHERENCE", source.Contract.ReadinessBPS, 10_000),
		indicator("gooo.metric.language.autonomous-change-proposal-ambiguous-predecessors.guardrail.v1",
			"GUARDRAIL", "REGRESSION", summary.AmbiguousCandidates, 0),
		indicator("gooo.metric.language.autonomous-change-proposal-unresolved.guardrail.v1",
			"GUARDRAIL", "REGRESSION", summary.Unresolved, 0),
		indicator("gooo.metric.language.autonomous-change-proposal-observer-writes.guardrail.v1",
			"GUARDRAIL", "FOUNDATION", summary.RepositoryWrites, 0),
		indicator("gooo.metric.language.autonomous-change-proposal-mutation-authority.guardrail.v1",
			"GUARDRAIL", "FOUNDATION", mutationAuthority(source), 0),
	}
}

func mutationAuthority(source Source) int {
	if source.Selection.SelectedPromotionAuthorized || source.Contract.PromotionAuthorized {
		return 1
	}
	return 0
}

func indicator(id, class, choice string, value, target int) Indicator {
	satisfied := value == target
	return Indicator{
		MetricID: id, Class: class, ProofChoice: choice,
		Producer: producer, Consumer: consumer, MetaOperation: operation,
		Value: value, Target: target, Satisfied: satisfied,
	}
}

func buildProofs(coordinates []Coordinate) []Proof {
	return []Proof{
		proof("FOUNDATION", "bind-merged-proposal-foundation", coordinates, 0, 1, 7),
		proof("COHERENCE", "cohere-proposal-selection-and-contract", coordinates, 2, 4, 5),
		proof("REGRESSION", "reject-ambiguous-or-writing-proposal", coordinates, 3, 6),
	}
}

func proof(choice, metaOperation string, coordinates []Coordinate, indexes ...int) Proof {
	selected, passed := make([]Coordinate, 0, len(indexes)), true
	for _, index := range indexes {
		selected = append(selected, coordinates[index])
		passed = passed && coordinates[index].Status == "SATISFIED"
	}
	return Proof{
		Choice: choice, MetaOperation: metaOperation, Passed: passed,
		EvidenceDigest: digestJSON(selected),
	}
}
