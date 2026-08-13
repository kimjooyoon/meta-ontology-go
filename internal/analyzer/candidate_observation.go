package analyzer

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const candidateObservationSchema = "analyzer-candidate-observation/v1"

const factObservationSchema = "analyzer-fact-observation/v1"

func sourceFactCanonical(fact Fact) string {
	var builder strings.Builder
	builder.WriteString(factObservationSchema)
	builder.WriteByte('\n')
	writeBindingField(&builder, fact.Subject.Namespace)
	writeBindingField(&builder, fact.Subject.ID)
	writeBindingField(&builder, string(fact.Relation))
	writeBindingField(&builder, fact.Object.Namespace)
	writeBindingField(&builder, fact.Object.ID)
	writeBindingField(&builder, string(fact.Origin))
	writeSemanticSpan(&builder, semanticSpan(fact.Span))
	return builder.String()
}

func validSourceFact(fact Fact) bool {
	if !fact.Subject.Valid() || !fact.Object.Valid() || !knownAnalyzerRelation(fact.Relation) ||
		fact.Span.Filename == "" || fact.Span.Start.Offset < 0 ||
		fact.Span.End.Offset < fact.Span.Start.Offset {
		return false
	}
	_, subjectErr := semantic.ParseIdentity(fact.Subject.ID)
	_, objectErr := semantic.ParseIdentity(fact.Object.ID)
	return subjectErr == nil && objectErr == nil
}

func candidateObservationDigest(candidate Candidate) string {
	var builder strings.Builder
	builder.WriteString(candidateObservationSchema)
	builder.WriteByte('\n')
	writeBindingField(&builder, candidate.Subject.Namespace)
	writeBindingField(&builder, candidate.Subject.ID)
	writeBindingField(&builder, string(candidate.Relation))
	writeBindingField(&builder, candidate.Reference)
	options := append([]Identity(nil), candidate.Options...)
	sort.Slice(options, func(i, j int) bool { return identityLess(options[i], options[j]) })
	for _, option := range options {
		writeBindingField(&builder, option.Namespace)
		writeBindingField(&builder, option.ID)
	}
	writeSemanticSpan(&builder, semanticSpan(candidate.Span))
	writeBindingField(&builder, candidate.Reason)
	writeBindingField(&builder, string(candidate.Origin))
	return semantic.StableHashString(builder.String())
}

func candidateObservationsMatch(result SemanticAdapterResult) bool {
	if len(result.DeferredCandidates) != len(result.NormalizedDelta.CandidateFacts) {
		return false
	}
	expected := make([]string, 0, len(result.DeferredCandidates))
	for _, candidate := range result.DeferredCandidates {
		expected = append(expected, candidateObservationDigest(candidate))
	}
	actual := make([]string, 0, len(result.NormalizedDelta.CandidateFacts))
	for _, candidate := range result.NormalizedDelta.CandidateFacts {
		actual = append(actual, candidate.ObservationDigest)
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
