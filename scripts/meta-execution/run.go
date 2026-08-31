package main

import (
	"fmt"
	"os"
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
	bundle, bundleErr := executeSelectedOperations(plan, manifest, workspaceRoot())
	if bundleErr != nil {
		return fmt.Errorf("execute selected operations: %w", bundleErr)
	}
	bundlePath := filepath.Join(filepath.Dir(configuration.planPath), "meta-operation-observations.json")
	bundlePayload, err := generation.EncodeObservationBundle(bundle)
	if err != nil {
		return fmt.Errorf("encode operation observations: %w", err)
	}
	if err := writeAtomic(bundlePath, bundlePayload); err != nil {
		return fmt.Errorf("write operation observations: %w", err)
	}
	printObservationSummary(bundle)
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

func printObservationSummary(bundle generation.OperationObservationBundle) {
	fmt.Printf(
		"operation observations: receipts=%d failures=%d total=%d replay=%d\n",
		len(bundle.Receipts),
		len(bundle.Failures),
		bundle.ObservationTotal,
		bundle.ReplayComparisons,
	)
	for _, failure := range bundle.Failures {
		fmt.Printf(
			"operation observation failure: action=%s decision=%s stage=%s step=%s reason=%s unknown_class=%s next_operation=%s blocked_by=%v counterexample=%s derived_relations=%v exit=%d stdout_bytes=%d raw_stdout_digest=%s stdout_digest=%s stderr_bytes=%d raw_stderr_digest=%s stderr_digest=%s\n",
			failure.ActionIndicatorID,
			failure.Decision,
			failure.Stage,
			failure.Step,
			failure.Reason,
			failure.UnknownClass,
			failure.NextOperation,
			failure.BlockedBy,
			failure.Counterexample,
			failure.DerivedRelations,
			failure.Executor.ExitCode,
			failure.Executor.StdoutBytes,
			failure.Executor.RawStdoutDigest,
			failure.Executor.StdoutDigest,
			failure.Executor.StderrBytes,
			failure.Executor.RawStderrDigest,
			failure.Executor.StderrDigest,
		)
		for _, evidence := range failure.FailureEvidence {
			fmt.Printf(
				"operation failure evidence: action=%s indicator_id=%s decision=%s observed=%d expected=%d counterexample=%s\n",
				failure.ActionIndicatorID,
				evidence.IndicatorID,
				evidence.Decision,
				evidence.Observed,
				evidence.Expected,
				evidence.Counterexample,
			)
		}
	}
}

func workspaceRoot() string {
	for _, name := range []string{"LOGICAL_WORKSPACE", "GITHUB_WORKSPACE"} {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	working, err := os.Getwd()
	if err != nil {
		return "."
	}
	return working
}
