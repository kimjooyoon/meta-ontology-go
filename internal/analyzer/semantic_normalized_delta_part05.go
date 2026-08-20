package analyzer

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func normalizedCandidateFacts(
	input SemanticAdapterInput, result SemanticAdapterResult, binding DeltaBinding,
) ([]NormalizedCandidateFact, error) {
	output := make([]NormalizedCandidateFact, 0, len(input.Analysis.Delta.Candidates))
	for _, sourceCandidate := range input.Analysis.Delta.Candidates {
		candidate := NormalizedCandidateFact{
			Binding: binding, ObservationDigest: candidateObservationDigest(sourceCandidate),
			SourceRelation: sourceCandidate.Relation,
			Origin:         sourceCandidate.Origin, Span: semanticSpan(sourceCandidate.Span),
			Reason: sourceCandidate.Reason,
		}
		subject, err := semantic.ParseIdentity(sourceCandidate.Subject.ID)
		if err != nil {
			return nil, err
		}
		candidate.Subject = subject
		for _, option := range sourceCandidate.Options {
			identity, err := semantic.ParseIdentity(option.ID)
			if err != nil {
				return nil, err
			}
			candidate.Options = append(candidate.Options, identity)
		}
		mapping, mapped := input.Policy.lookup(sourceCandidate.Relation)
		if !mapped || !mapping.allowsOrigin(sourceCandidate.Origin) {
			slices.Sort(candidate.Options)
			output = append(output, candidate)
			continue
		}
		for _, option := range sourceCandidate.Options {
			fact, err := mapFact(result.IR.Graph, sourceCandidate.Subject, option, mapping, sourceCandidate.Span)
			if err != nil {
				continue
			}
			fact.Status = semantic.FactCandidate
			fact.Reason = sourceCandidate.Reason
			candidate.Facts = append(candidate.Facts, fact)
			if evidence, ok := evidenceForFact(result.IR.Evidence(), fact.Key(), semantic.FactCandidate); ok {
				candidate.Evidence = append(candidate.Evidence, evidence)
			} else if evidence, ok := shadowEvidenceForFact(result.ShadowedCandidateEvidence, fact.Key()); ok {
				candidate.Evidence = append(candidate.Evidence, evidence)
			}
		}
		sortNormalizedCandidate(&candidate)
		output = append(output, candidate)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].canonical() < output[j].canonical() })
	return output, nil
}
func sortNormalizedCandidate(candidate *NormalizedCandidateFact) {
	slices.Sort(candidate.Options)
	sort.Slice(candidate.Facts, func(i, j int) bool {
		return candidate.Facts[i].Canonical() < candidate.Facts[j].Canonical()
	})
	sort.Slice(candidate.Evidence, func(i, j int) bool {
		return candidate.Evidence[i].Canonical() < candidate.Evidence[j].Canonical()
	})
}
