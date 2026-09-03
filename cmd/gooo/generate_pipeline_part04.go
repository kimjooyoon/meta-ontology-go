package main

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"time"
)

func generateWithDeadline(file *syntax.File, previous []byte, timeout time.Duration) (generationResult, error) {
	return generateWithDeadlineObserved(file, previous, timeout, nil)
}

func generateWithDeadlineObserved(file *syntax.File, previous []byte, timeout time.Duration, recorder *generation.SemanticObservationRecorder) (generationResult, error) {
	if timeout <= 0 {
		return generationResult{}, errCommandDeadline
	}
	result := make(chan generationResult, 1)
	go func() {
		ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV1Support())
		if err != nil {
			result <- generationResult{err: fmt.Errorf("semantic lowering failed: %w", err)}
			return
		}
		if recorder != nil {
			digest, err := cache.SemanticDigest(ir)
			if err != nil {
				result <- generationResult{err: fmt.Errorf("semantic observation input digest failed: %w", err)}
				return
			}
			if err := recorder.Record(generation.SemanticObservationEvent{
				Phase:       generation.SemanticObservationPhase,
				OperationID: generation.SemanticObservationOperationID,
				InputDigest: digest.String(),
				Pure:        true,
				Effects:     []string{"read:source", "read:semantic-ir"},
				SourceSpans: []generation.SemanticObservationSpan{semanticObservationSpan(file)},
			}); err != nil {
				result <- generationResult{err: fmt.Errorf("semantic observation record failed: %w", err)}
				return
			}
		}
		model, err := projectionIR(ir)
		if err != nil {
			result <- generationResult{err: fmt.Errorf("generator adapter failed: %w", err)}
			return
		}
		if semanticIRHasFields(ir) {
			document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV1Support())
			if err != nil {
				result <- generationResult{err: fmt.Errorf("BX document adaptation failed: %w", err)}
				return
			}
			sourceModel, err := bidir.GetWithEntityFieldsSupport(document, syntax.EntityFieldsV1Support())
			if err != nil {
				result <- generationResult{err: fmt.Errorf("BX model projection failed: %w", err)}
				return
			}
			model, err = projectionIRFromBidirModelWithSupport(ir, sourceModel, syntax.EntityFieldsV1Support())
			if err != nil {
				result <- generationResult{err: fmt.Errorf("generator field adapter failed: %w", err)}
				return
			}
		}
		var generated generator.Result
		if semanticIRHasFields(ir) {
			generated, err = generator.GenerateWithEntityFieldsSupport(model, previous, syntax.EntityFieldsV1Support())
		} else {
			generated, err = generator.Generate(model, previous)
		}
		result <- generationResult{ir: ir, result: generated, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case generated := <-result:
		return generated, generated.err
	case <-timer.C:
		return generationResult{}, errCommandDeadline
	}
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
func generateSource(file *syntax.File) ([]byte, error) {
	result, err := generateWithDeadline(file, nil, commandDeadline)
	if err != nil {
		return nil, err
	}
	return result.result.Source, nil
}
