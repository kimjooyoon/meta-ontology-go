package main

import (
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func semanticCheckIR(file *syntax.File, timeout time.Duration) (semantic.IR, error) {
	ir, err := lowerInspectIR(file, timeout)
	if err != nil {
		return semantic.IR{}, err
	}
	if err := ir.Validate(); err != nil {
		return semantic.IR{}, err
	}
	return ir, nil
}
