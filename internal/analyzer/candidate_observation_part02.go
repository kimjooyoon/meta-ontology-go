package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func shadowedCandidateEvidenceMatch(result SemanticAdapterResult) bool {
	expected := make([]string, 0)
	for _, candidate := range result.NormalizedDelta.CandidateFacts {
		for _, fact := range candidate.Facts {
			if result.IR.Graph.HasCandidate(fact.Key()) {
				continue
			}
			evidence, ok := evidenceForFact(candidate.Evidence, fact.Key(), semantic.FactCandidate)
			if !ok {
				return false
			}
			expected = append(expected, evidence.Canonical())
		}
	}
	actual := make([]string, 0, len(result.ShadowedCandidateEvidence))
	for _, evidence := range result.ShadowedCandidateEvidence {
		if evidence.Status != semantic.FactCandidate {
			return false
		}
		actual = append(actual, evidence.Canonical())
	}
	sort.Strings(expected)
	sort.Strings(actual)
	if len(expected) != len(actual) {
		return false
	}
	for index := range expected {
		if expected[index] != actual[index] {
			return false
		}
	}
	return true
}
func deferredFactsMatch(result SemanticAdapterResult) bool {
	if len(result.DeferredFacts) != len(result.NormalizedDelta.DeferredFacts) {
		return false
	}
	expected := make([]string, 0, len(result.DeferredFacts))
	for _, fact := range result.DeferredFacts {
		expected = append(expected, sourceFactCanonical(fact))
	}
	actual := make([]string, 0, len(result.NormalizedDelta.DeferredFacts))
	for _, fact := range result.NormalizedDelta.DeferredFacts {
		actual = append(actual, sourceFactCanonical(fact.Fact))
	}
	sort.Strings(expected)
	sort.Strings(actual)
	for index := range expected {
		if expected[index] != actual[index] {
			return false
		}
	}
	return true
}
