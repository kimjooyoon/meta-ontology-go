package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func validImplementationObservation(observation ImplementationObservation) bool {
	if !validDigest(observation.SourceDigest) || !validDigest(observation.BaseDigest) ||
		!validDigest(observation.PolicyDigest) || !validDigest(observation.ToolchainDigest) ||
		!validDigest(observation.RegistryDigest) || observation.SourceFile == "" ||
		observation.Span.Filename != observation.SourceFile ||
		observation.Span.Start.Offset < 0 ||
		observation.Span.End.Offset < observation.Span.Start.Offset ||
		observation.Origin != OriginImplementation || !knownAnalyzerRelation(observation.Relation) ||
		!observation.Subject.Valid() || !observation.Object.Valid() {
		return false
	}
	if _, err := semantic.ParseIdentity(observation.Subject.ID); err != nil {
		return false
	}
	if _, err := semantic.ParseIdentity(observation.Object.ID); err != nil {
		return false
	}
	return true
}
func collectImplementationObservations(
	result Result, base semantic.IR, input SemanticAdapterInput,
) []ImplementationObservation {
	observations := make([]ImplementationObservation, 0)
	for _, fact := range result.Delta.Added {
		if fact.Origin != OriginImplementation {
			continue
		}
		mapping, mapped := input.Policy.lookup(fact.Relation)
		if mapped && mapping.allowsOrigin(fact.Origin) {
			continue
		}
		observations = append(observations, ImplementationObservation{
			SourceDigest: input.SourceDigest, SourceFile: fact.Span.Filename,
			BaseDigest: base.StableHash(), PolicyDigest: input.Policy.Digest(),
			ToolchainDigest: input.ToolchainDigest, RegistryDigest: input.Registry.Digest(), Subject: fact.Subject,
			Relation: fact.Relation, Object: fact.Object, Origin: fact.Origin,
			Span: fact.Span,
		})
	}
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].Canonical() < observations[j].Canonical()
	})
	return observations
}
func implementationObservationsMatch(
	left, right []ImplementationObservation,
) bool {
	orderedLeft := append([]ImplementationObservation(nil), left...)
	orderedRight := append([]ImplementationObservation(nil), right...)
	sort.Slice(orderedLeft, func(i, j int) bool {
		return orderedLeft[i].Canonical() < orderedLeft[j].Canonical()
	})
	sort.Slice(orderedRight, func(i, j int) bool {
		return orderedRight[i].Canonical() < orderedRight[j].Canonical()
	})
	if len(orderedLeft) != len(orderedRight) {
		return false
	}
	for index := range orderedLeft {
		if orderedLeft[index].Canonical() != orderedRight[index].Canonical() {
			return false
		}
	}
	return true
}
