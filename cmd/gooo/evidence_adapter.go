package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// CompilerEvidenceEmitter is the CLI seam for a future bootstrap runner. It
// emits fact-level evidence without deciding whether a compiler stage passed.
// Stage verdicts belong to the independent bootstrap verifier.
type CompilerEvidenceEmitter interface {
	EmitCompilerEvidence(semantic.IR, semantic.ID) ([]semantic.Evidence, error)
}

// SemanticEvidenceEmitter creates deterministic compiler-run evidence for
// either the Go-hosted or gooo-hosted compiler identity.
type SemanticEvidenceEmitter struct{}

func (SemanticEvidenceEmitter) EmitCompilerEvidence(ir semantic.IR, producer semantic.ID) ([]semantic.Evidence, error) {
	if producer != semantic.GoHostedCompilerID && producer != semantic.GoooHostedCompilerID {
		return nil, fmt.Errorf("unsupported compiler evidence producer %q", producer)
	}
	if err := ir.Validate(); err != nil {
		return nil, err
	}
	facts := ir.Graph.AllFacts()
	evidence := make([]semantic.Evidence, 0, len(facts))
	for _, fact := range facts {
		record, err := compilerEvidence(fact, producer)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, record)
	}
	return evidence, nil
}

func compilerEvidence(fact semantic.Fact, producer semantic.ID) (semantic.Evidence, error) {
	evidenceID := semantic.ID("gooo://evidence/fact/" + semantic.StableHashString(fact.SemanticCanonical()))
	record := semantic.Evidence{
		ID:       evidenceID,
		Producer: producer,
		Kind:     semantic.CompilerRunEvidence,
		Fact:     fact.Key(),
		Status:   fact.Status,
		Digest:   fact.StableHash(),
		Span:     fact.Span,
	}
	return record.Normalized()
}
