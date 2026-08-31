package replay

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"os"
)

func Observe(root, relativePath string) (Observation, error) {
	target, cleanPath, err := sourcePath(root, relativePath)
	if err != nil {
		return Observation{}, err
	}
	source, err := os.ReadFile(target)
	if err != nil {
		return Observation{}, fmt.Errorf("read %s: %w", cleanPath, err)
	}
	first, err := lower(cleanPath, source)
	if err != nil {
		return Observation{}, err
	}
	second, err := lower(cleanPath, source)
	if err != nil {
		return Observation{}, fmt.Errorf("replay %s: %w", cleanPath, err)
	}
	comparison := semantic.CompareIR(first, second)
	return Observation{
		Path:               cleanPath,
		SourceLines:        physicalLines(source),
		SourceDigest:       semantic.StableHash(source),
		IRVersion:          first.Version,
		Package:            first.Package,
		Namespace:          first.Namespace.String(),
		Nodes:              len(first.Graph.Nodes()),
		DeterministicFacts: len(first.Graph.DeterministicFacts()),
		CandidateFacts:     len(first.Graph.Candidates()),
		Normalized:         first.Validate() == nil,
		CanonicalReplay:    first.Canonical() == second.Canonical(),
		SemanticReplay:     comparison.SemanticEqual,
		ProvenanceReplay:   comparison.ProvenanceEqual,
		EvidenceReplay:     comparison.ExactEvidenceEqual,
		SemanticHash:       first.StableHash(),
		ProvenanceHash:     first.ProvenanceHash(),
		EvidenceHash:       first.EvidenceHash(),
		Stages:             append([]string(nil), ExpectedStages...),
		Effects: EffectReceipt{
			Reads: []string{cleanPath},
		},
		IR: first,
	}, nil
}
