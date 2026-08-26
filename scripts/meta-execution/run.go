package main

import (
	"fmt"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type options struct {
	planPath   string
	outputPath string
}

func run(configuration options) error {
	if configuration.planPath == "" || configuration.outputPath == "" ||
		filepath.Clean(configuration.planPath) ==
			filepath.Clean(configuration.outputPath) {
		return fmt.Errorf("plan and output must be distinct non-empty paths")
	}
	plan := generation.Plan{}
	if err := decodeJSON(configuration.planPath, &plan); err != nil {
		return err
	}
	manifest := generation.BuildExecutionManifest(plan)
	payload, err := generation.EncodeExecutionManifest(manifest)
	if err != nil {
		return fmt.Errorf("encode execution manifest: %w", err)
	}
	if err := writeAtomic(configuration.outputPath, payload); err != nil {
		return err
	}
	fmt.Printf(
		"execution manifest: decision=%s reason=%s replay=%s\n",
		manifest.Decision,
		manifest.Reason,
		manifest.ReplayDigest,
	)
	if manifest.Decision != generation.ExecutionDecisionFixedPoint &&
		manifest.Decision != generation.ExecutionDecisionProposed {
		return fmt.Errorf(
			"execution manifest failed: %s/%s",
			manifest.Decision,
			manifest.Reason,
		)
	}
	return nil
}
