package replay

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func ObserveLaws(anchorPath string, input semantic.IR) (LawObservation, error) {
	base, err := input.Normalized()
	if err != nil {
		return LawObservation{}, err
	}
	nodes := base.Graph.Nodes()
	facts := base.Graph.DeterministicFacts()
	if len(nodes) == 0 {
		return LawObservation{}, fmt.Errorf("semantic law anchor %s has no nodes", anchorPath)
	}
	if len(facts) == 0 {
		return LawObservation{}, fmt.Errorf("semantic law anchor %s has no deterministic facts", anchorPath)
	}

	presentation, err := base.Normalized()
	if err != nil {
		return LawObservation{}, err
	}
	presented := nodes[0]
	presented.Name += " semantic presentation"
	if err := presentation.Graph.AddNode(presented); err != nil {
		return LawObservation{}, err
	}
	if err := presentation.Validate(); err != nil {
		return LawObservation{}, err
	}

	structure, err := structureOnly(base)
	if err != nil {
		return LawObservation{}, err
	}
	candidateIR, err := structure.Normalized()
	if err != nil {
		return LawObservation{}, err
	}
	candidate := facts[0]
	candidate.Reason = "semantic-model-candidate-law"
	if err := candidateIR.AddCandidate(candidate); err != nil {
		return LawObservation{}, err
	}
	if err := candidateIR.Validate(); err != nil {
		return LawObservation{}, err
	}
	deterministicIR, err := structure.Normalized()
	if err != nil {
		return LawObservation{}, err
	}
	if err := deterministicIR.AddFact(facts[0]); err != nil {
		return LawObservation{}, err
	}
	if err := deterministicIR.Validate(); err != nil {
		return LawObservation{}, err
	}

	return LawObservation{
		AnchorPath:                    anchorPath,
		PresentationChanged:           base.Canonical() != presentation.Canonical(),
		PresentationInvariant:         base.StableHash() == presentation.StableHash(),
		CandidateRecorded:             len(candidateIR.Graph.Candidates()) == 1,
		CandidateNonAuthoritative:     structure.StableHash() == candidateIR.StableHash(),
		DeterministicRecorded:         len(deterministicIR.Graph.DeterministicFacts()) == 1,
		DeterministicAuthoritative:    structure.StableHash() != deterministicIR.StableHash(),
		StructureSemanticHash:         structure.StableHash(),
		PresentationSemanticHash:      presentation.StableHash(),
		CandidateSemanticHash:         candidateIR.StableHash(),
		DeterministicSemanticHash:     deterministicIR.StableHash(),
		CandidateCanonicalChanged:     structure.Canonical() != candidateIR.Canonical(),
		DeterministicCanonicalChanged: structure.Canonical() != deterministicIR.Canonical(),
	}, nil
}
