package main

import (
	"fmt"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type options struct {
	planPath     string
	receiptsPath string
	outputPath   string
}

func run(configuration options) error {
	if !optionsKnown(configuration) {
		return fmt.Errorf("plan and output must be distinct non-empty paths")
	}
	plan := generation.Plan{}
	if err := decodeJSON(configuration.planPath, &plan); err != nil {
		return err
	}
	receipts := []generation.OperationReceipt{}
	if configuration.receiptsPath != "" {
		if err := decodeJSON(configuration.receiptsPath, &receipts); err != nil {
			return err
		}
	} else {
		manifestPath := filepath.Join(filepath.Dir(configuration.planPath), "self-improvement-execution.json")
		manifest := generation.ExecutionManifest{}
		if err := decodeJSON(manifestPath, &manifest); err != nil {
			return fmt.Errorf("read execution manifest: %w", err)
		}
		bundlePath := filepath.Join(filepath.Dir(configuration.planPath), "meta-operation-observations.json")
		bundle := generation.OperationObservationBundle{}
		if err := decodeJSON(bundlePath, &bundle); err != nil {
			return fmt.Errorf("read operation observations: %w", err)
		}
		if err := generation.ValidateObservationBundle(bundle, plan, manifest); err != nil {
			return fmt.Errorf("operation observation binding failed: %w", err)
		}
		receipts = bundle.Receipts
	}
	report := generation.VerifyReceipts(plan, receipts)
	payload, err := generation.EncodeReceiptReport(report)
	if err != nil {
		return fmt.Errorf("encode receipt report: %w", err)
	}
	if err := writeAtomic(configuration.outputPath, payload); err != nil {
		return err
	}
	fmt.Printf(
		"receipt verification: decision=%s reason=%s unknown=%d replay=%s\n",
		report.Decision,
		report.Reason,
		len(report.Unknowns),
		report.ReplayDigest,
	)
	if report.Decision != generation.ReceiptDecisionFixedPoint &&
		report.Decision != generation.ReceiptDecisionConformant {
		return fmt.Errorf(
			"receipt verification failed: %s/%s",
			report.Decision,
			report.Reason,
		)
	}
	return nil
}

func optionsKnown(configuration options) bool {
	if configuration.planPath == "" || configuration.outputPath == "" {
		return false
	}
	output := filepath.Clean(configuration.outputPath)
	if output == filepath.Clean(configuration.planPath) {
		return false
	}
	return configuration.receiptsPath == "" ||
		output != filepath.Clean(configuration.receiptsPath)
}
