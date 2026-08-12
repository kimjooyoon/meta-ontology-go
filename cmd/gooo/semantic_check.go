package main

import (
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const semanticCheckSchemaVersion = "gooo-semantic-check/v1"

func semanticCheckIR(file *syntax.File, timeout time.Duration) (semantic.IR, error) {
	return semanticCheckIRWithLowerer(file, timeout, bidir.Lower)
}

func semanticCheckIRWithLowerer(file *syntax.File, timeout time.Duration, lower func(*syntax.File) (semantic.IR, error)) (semantic.IR, error) {
	ir, err := lowerInspectIRWith(file, timeout, lower)
	if err != nil {
		return semantic.IR{}, err
	}
	if err := ir.Validate(); err != nil {
		return semantic.IR{}, err
	}
	return ir, nil
}
