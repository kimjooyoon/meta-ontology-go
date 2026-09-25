package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func normalizedDeferredFacts(facts []Fact, binding DeltaBinding) []NormalizedDeferredFact {
	output := make([]NormalizedDeferredFact, 0, len(facts))
	for _, fact := range facts {
		output = append(output, NormalizedDeferredFact{Binding: binding, Fact: fact})
	}
	sort.Slice(output, func(i, j int) bool { return output[i].canonical() < output[j].canonical() })
	return output
}
func normalizedSignatureFacts(
	input SemanticAdapterInput, result SemanticAdapterResult, binding DeltaBinding,
) []NormalizedSignatureFact {
	output := make([]NormalizedSignatureFact, 0)
	for _, sourceFact := range input.Analysis.Delta.Added {
		if sourceFact.Origin != OriginSignature {
			continue
		}
		mapping, ok := input.Policy.lookup(sourceFact.Relation)
		if !ok || !mapping.allowsOrigin(sourceFact.Origin) {
			continue
		}
		mapped, err := mapFact(result.IR.Graph, sourceFact.Subject, sourceFact.Object, mapping, sourceFact.Span)
		if err != nil {
			continue
		}
		evidence, ok := evidenceForFact(result.IR.Evidence(), mapped.Key(), semantic.FactDeterministic)
		if !ok {
			continue
		}
		output = append(output, NormalizedSignatureFact{
			Binding: binding, SourceRelation: sourceFact.Relation, Fact: mapped, Evidence: evidence,
		})
	}
	sort.Slice(output, func(i, j int) bool { return output[i].canonical() < output[j].canonical() })
	return output
}
