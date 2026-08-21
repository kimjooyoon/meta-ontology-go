package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateObservations(result Result) error {
	for _, fact := range result.Delta.Added {
		if !knownAnalyzerRelation(fact.Relation) {
			return adapterError(AdapterUnknownRelation, fact.Relation, "", "fact relation is not analyzer vocabulary")
		}
	}
	for _, candidate := range result.Delta.Candidates {
		if !knownAnalyzerRelation(candidate.Relation) {
			return adapterError(AdapterUnknownRelation, candidate.Relation, "", "candidate relation is not analyzer vocabulary")
		}
	}
	return nil
}
func hasMappedObservation(result Result, policy MappingPolicy) bool {
	for _, fact := range result.Delta.Added {
		if _, ok := policy.lookup(fact.Relation); ok {
			return true
		}
	}
	for _, candidate := range result.Delta.Candidates {
		if _, ok := policy.lookup(candidate.Relation); ok && len(candidate.Options) > 0 {
			return true
		}
	}
	return false
}
func adaptFacts(result *SemanticAdapterResult, input SemanticAdapterInput) error {
	for _, fact := range input.Analysis.Delta.Added {
		mapping, ok := input.Policy.lookup(fact.Relation)
		if !ok {
			result.DeferredFacts = append(result.DeferredFacts, fact)
			continue
		}
		if !mapping.allowsOrigin(fact.Origin) {
			result.DeferredFacts = append(result.DeferredFacts, fact)
			continue
		}
		mapped, err := mapFact(result.IR.Graph, fact.Subject, fact.Object, mapping, fact.Span)
		if err != nil {
			return err
		}
		if err := result.IR.AddFact(mapped); err != nil {
			return err
		}
		evidence, err := mappedEvidence(input, fact.Relation, mapped, semantic.FactDeterministic)
		if err != nil {
			return err
		}
		if err := result.IR.AddEvidence(evidence); err != nil {
			return err
		}
	}
	return nil
}
