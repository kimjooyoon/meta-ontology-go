package analyzer

import (
	"encoding/hex"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

func adaptCandidates(result *SemanticAdapterResult, input SemanticAdapterInput) error {
	for _, candidate := range input.Analysis.Delta.Candidates {
		mapping, ok := input.Policy.lookup(candidate.Relation)
		if !ok || !mapping.allowsOrigin(candidate.Origin) {
			continue
		}
		for _, option := range candidate.Options {
			mapped, err := mapFact(result.IR.Graph, candidate.Subject, option, mapping, candidate.Span)
			if err != nil {
				return err
			}
			mapped.Status = semantic.FactCandidate
			mapped.Reason = candidate.Reason
			evidence, err := mappedEvidence(input, candidate.Relation, mapped, semantic.FactCandidate)
			if err != nil {
				return err
			}
			if result.IR.Graph.HasFact(mapped.Key()) {
				result.ShadowedCandidateEvidence = append(result.ShadowedCandidateEvidence, evidence)
				continue
			}
			if err := result.IR.AddCandidate(mapped); err != nil {
				return err
			}
			if err := result.IR.AddEvidence(evidence); err != nil {
				return err
			}
		}
	}
	sort.Slice(result.ShadowedCandidateEvidence, func(i, j int) bool {
		return result.ShadowedCandidateEvidence[i].Canonical() < result.ShadowedCandidateEvidence[j].Canonical()
	})
	return nil
}
func copyCandidates(candidates []Candidate) []Candidate {
	copyOf := append([]Candidate(nil), candidates...)
	for index := range copyOf {
		copyOf[index].Options = append([]Identity(nil), copyOf[index].Options...)
	}
	return copyOf
}
func copyDetails(details []ImplementationDetail) []ImplementationDetail {
	copyOf := append([]ImplementationDetail(nil), details...)
	for index := range copyOf {
		copyOf[index] = copyOf[index].normalized()
	}
	return copyOf
}
func validateEvidenceConfig(input SemanticAdapterInput) error {
	if _, err := semantic.ParseIdentity(input.Producer.String()); err != nil {
		return adapterError(AdapterEvidenceConfig, "", input.Producer.String(), "producer is not a semantic identity")
	}
	if !input.EvidenceKind.Valid() {
		return adapterError(AdapterEvidenceConfig, "", "", "evidence kind is invalid")
	}
	digest := strings.TrimSpace(input.SourceDigest)
	if len(digest) != 64 {
		return adapterError(AdapterEvidenceConfig, "", "", "source digest must be SHA-256")
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return adapterError(AdapterEvidenceConfig, "", "", "source digest is not hexadecimal")
	}
	return nil
}
