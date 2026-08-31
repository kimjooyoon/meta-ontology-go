package main

import (
	"context"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const semanticCheckSchemaVersion = "gooo-semantic-check/v1"

func semanticCheckIR(file *syntax.File, timeout time.Duration) (semantic.IR, error) {
	return semanticCheckIRWithLowerer(file, timeout, bidir.Lower)
}

func semanticCheckIRWithEntityFieldsSupport(file *syntax.File, timeout time.Duration) (semantic.IR, error) {
	return lowerEntityFieldsInspectIRWith(file, timeout, func(file *syntax.File) (semantic.IR, error) {
		return bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV1Support())
	})
}

func lowerEntityFieldsInspectIRWith(file *syntax.File, timeout time.Duration, lower func(*syntax.File) (semantic.IR, error)) (semantic.IR, error) {
	if timeout <= 0 {
		return semantic.IR{}, errCommandDeadline
	}
	result := make(chan inspectLowerResult, 1)
	go func() {
		ir, err := lower(file)
		if err == nil {
			err = ir.Validate()
		}
		result <- inspectLowerResult{ir: ir, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case lowered := <-result:
		return lowered.ir, lowered.err
	case <-timer.C:
		return semantic.IR{}, errCommandDeadline
	}
}

func semanticCheckIRWithLowerer(file *syntax.File, timeout time.Duration, lower func(*syntax.File) (semantic.IR, error)) (semantic.IR, error) {
	return lowerInspectIRWith(file, timeout, lower)
}
