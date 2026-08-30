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
	return semanticCheckIRWithLowerer(file, timeout, func(file *syntax.File) (semantic.IR, error) {
		return bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV1Support())
	})
}

func semanticCheckIRWithLowerer(file *syntax.File, timeout time.Duration, lower func(*syntax.File) (semantic.IR, error)) (semantic.IR, error) {
	return lowerInspectIRWith(file, timeout, lower)
}
