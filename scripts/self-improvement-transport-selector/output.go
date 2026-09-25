package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func writeReceipt(path string, receipt selfimprovementtransport.LifecycleReceipt) error {
	if err := selfimprovementtransport.ValidateArtifactLifecycleReceipt(receipt); err != nil {
		return err
	}
	return writeJSON(path, receipt)
}

func writeJSON(path string, value any) error {
	if path == "" {
		return fmt.Errorf("output path is required")
	}
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode output: %w", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func printReceipt(receipt selfimprovementtransport.LifecycleReceipt) {
	fmt.Printf("artifact lifecycle: %s %d/%d (%d bps), unknown path %d at %s/%s\n",
		receipt.Decision, receipt.Metrics.VerifiedTotal, receipt.Metrics.FixedStepTotal,
		receipt.Metrics.CoverageBasisPoints, receipt.Metrics.UnknownPathCount,
		receipt.Coordinate.Stage, receipt.Coordinate.Step)
}
