package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure"
)

func readInput(value config) (closure.Input, error) {
	program, err := os.ReadFile(value.programPath)
	if err != nil {
		return closure.Input{}, fmt.Errorf("read program: %w", err)
	}
	source, err := os.ReadFile(value.sourcePath)
	if err != nil {
		return closure.Input{}, fmt.Errorf("read source: %w", err)
	}
	verification, err := os.ReadFile(value.verificationPath)
	if err != nil {
		return closure.Input{}, fmt.Errorf("read verification: %w", err)
	}
	return newInput(value, program, source, verification), nil
}
