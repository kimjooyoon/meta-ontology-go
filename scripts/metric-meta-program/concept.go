package main

import (
	"fmt"

	conceptoperation "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func emitConceptReceipt(directory, interventionPath string,
	strategy, strategyVerification, program, source, report []byte) error {
	intervention, err := readBounded(interventionPath)
	if err != nil {
		return fmt.Errorf("read intervention: %w", err)
	}
	receipt, err := conceptoperation.Build(strategy, strategyVerification, intervention, program, source, report)
	if err != nil {
		return fmt.Errorf("build concept-operation receipt: %w", err)
	}
	payload, err := encodeJSON(receipt)
	if err != nil {
		return fmt.Errorf("encode concept-operation receipt: %w", err)
	}
	if err := writeConceptOperationReceipt(directory, payload); err != nil {
		return fmt.Errorf("write concept-operation receipt: %w", err)
	}
	return nil
}
