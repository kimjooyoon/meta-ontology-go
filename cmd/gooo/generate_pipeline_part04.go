package main

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"time"
)

func generateWithDeadline(file *syntax.File, previous []byte, timeout time.Duration) (generationResult, error) {
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
func generateSource(file *syntax.File) ([]byte, error) {
	result, err := generateWithDeadline(file, nil, commandDeadline)
	if err != nil {
		return nil, err
	}
	return result.result.Source, nil
}
