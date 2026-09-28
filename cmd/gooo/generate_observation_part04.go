package main

import (
	"fmt"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func generateWithDeadline(file *syntax.File, previous []byte, timeout time.Duration) (generationResult, error) {
	return generateWithDeadlineObserved(file, previous, timeout, nil)
}

func generateWithDeadlineObserved(file *syntax.File, previous []byte, timeout time.Duration, recorder *generation.SemanticObservationRecorder) (generationResult, error) {
	result, err := generateWithDeadlineCore(file, previous, timeout)
	if err != nil || recorder == nil {
		return result, err
	}
	digest, err := cache.SemanticDigest(result.ir)
	if err != nil {
		return generationResult{}, fmt.Errorf("semantic observation input digest failed: %w", err)
	}
	if err := recorder.Record(generation.SemanticObservationEvent{
		Phase:       generation.SemanticObservationPhase,
		OperationID: generation.SemanticObservationOperationID,
		InputDigest: digest.String(),
		Pure:        true,
		Effects:     []string{"read:source", "read:semantic-ir"},
		SourceSpans: []generation.SemanticObservationSpan{semanticObservationSpan(file)},
	}); err != nil {
		return generationResult{}, fmt.Errorf("semantic observation record failed: %w", err)
	}
	return result, nil
}

func semanticObservationSpan(file *syntax.File) generation.SemanticObservationSpan {
	span := syntaxFileSpan(file)
	return generation.SemanticObservationSpan{
		File:        span.Filename,
		StartOffset: span.Start.Offset,
		StartLine:   span.Start.Line,
		StartColumn: span.Start.Column,
		EndOffset:   span.End.Offset,
		EndLine:     span.End.Line,
		EndColumn:   span.End.Column,
	}
}
