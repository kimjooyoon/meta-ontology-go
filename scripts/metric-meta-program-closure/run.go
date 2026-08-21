package main

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure"
	closureverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/closure/verify"
)

func run(value config) error {
	input, err := readInput(value)
	if err != nil {
		return err
	}
	receipt, err := closure.Build(input)
	if err != nil {
		return fmt.Errorf("build closure: %w", err)
	}
	receiptJSON, err := marshal(receipt)
	if err != nil {
		return err
	}
	report, err := closureverify.Verify(verifierInput(input), receiptJSON)
	if err != nil {
		return fmt.Errorf("verify closure: %w", err)
	}
	reportJSON, err := marshal(report)
	if err != nil {
		return err
	}
	if err := writeOutputs(value.outputDir, receiptJSON, reportJSON); err != nil {
		return err
	}
	fmt.Printf("verified: %s artifact=%d files=3
", report.ReceiptDigest, report.ArtifactID)
	return nil
}

func marshal(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	return append(raw, '
'), err
}
