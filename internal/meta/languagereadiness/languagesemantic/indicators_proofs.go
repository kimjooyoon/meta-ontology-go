package languagesemantic

import (
	"fmt"
)

func buildProofs(summary Summary, source Source, resolution Resolution) []Proof {
	values := []struct {
		choice, operation, evidence string
		passed                      bool
	}{
		{"FOUNDATION", "bind-versioned-semantic-corpus", fmt.Sprintf("%s|%s|%d|%d|%d", source.RegistryDigest, source.SyntaxArtifactDigest, summary.UnregisteredGooo, summary.MissingRegistered, summary.RegistryDrift), source.ObservationKnown && source.ConceptBound && summary.UnregisteredGooo == 0 && summary.MissingRegistered == 0 && summary.RegistryDrift == 0},
		{"COHERENCE", "replay-normalized-authoritative-meaning", fmt.Sprintf("%d|%d|%d|%d|%d|%d", summary.NormalizedIRs, summary.SemanticReplays, summary.ProvenanceReplays, summary.EvidenceReplays, summary.PresentationLaws, summary.DeterministicAuthorityLaws), summary.NormalizedIRs == expectedSources && summary.SemanticReplays == expectedSources && summary.ProvenanceReplays == expectedSources && summary.EvidenceReplays == expectedSources && summary.PresentationLaws == 1 && summary.DeterministicAuthorityLaws == 1},
		{"REGRESSION", "reject-unknown-effects-and-candidate-authority", fmt.Sprintf("%d|%d|%d|%d", summary.UpstreamRejections, summary.CandidateAuthorityLaws, summary.EffectfulStages, summary.Unresolved), summary.UpstreamRejections == expectedRejections && summary.CandidateAuthorityLaws == 1 && summary.EffectfulStages == 0 && summary.Unresolved == 0 && resolution == ResolutionExact},
	}
	proofs := make([]Proof, 0, len(values))
	for _, value := range values {
		proofs = append(proofs, Proof{Choice: value.choice, MetaOperation: value.operation, EvidenceDigest: semanticHash(value.evidence), Passed: value.passed})
	}
	return proofs
}
