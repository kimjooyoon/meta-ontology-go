package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type options struct {
	planPath      string
	executionPath string
	receiptsPath  string
	outputPath    string
}

func run(configuration options) error {
	if !optionsKnown(configuration) {
		return fmt.Errorf("plan, execution, receipts, and output must be distinct paths")
	}
	plan := generation.Plan{}
	execution := generation.ExecutionManifest{}
	receipts := generation.ReceiptReport{}
	if err := decodeJSON(configuration.planPath, &plan); err != nil {
		return err
	}
	if err := decodeJSON(configuration.executionPath, &execution); err != nil {
		return err
	}
	if err := decodeJSON(configuration.receiptsPath, &receipts); err != nil {
		return err
	}
	envelope := generation.BindArtifactProvenance(plan, execution, receipts)
	payload, err := generation.EncodeArtifactProvenance(envelope)
	if err != nil {
		return fmt.Errorf("encode artifact provenance: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(configuration.outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(configuration.outputPath, payload, 0o644); err != nil {
		return fmt.Errorf("write artifact provenance: %w", err)
	}
	fmt.Printf("artifact provenance: decision=%s reason=%s replay=%s\n",
		envelope.Decision, envelope.Reason, envelope.ReplayDigest)
	if envelope.Decision != generation.ArtifactProvenanceDecisionBound {
		return fmt.Errorf("artifact provenance failed: %s/%s",
			envelope.Decision, envelope.Reason)
	}
	return nil
}

func optionsKnown(configuration options) bool {
	paths := []string{configuration.planPath, configuration.executionPath,
		configuration.receiptsPath, configuration.outputPath}
	known := map[string]struct{}{}
	for _, path := range paths {
		if path == "" {
			return false
		}
		clean := filepath.Clean(path)
		if _, exists := known[clean]; exists {
			return false
		}
		known[clean] = struct{}{}
	}
	return true
}
