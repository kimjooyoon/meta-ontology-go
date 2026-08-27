package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

type options struct {
	mode, contract, runPath, artifactsPath, repository string
	artifactName, output, receipt, archive             string
	runID                                              int64
	runAttempt, runLookupExit, artifactsLookupExit     int
	downloadExit                                       int
}

func run(input options) error {
	switch input.mode {
	case "locate":
		return runLocate(input)
	case "transport":
		return runTransport(input)
	default:
		return fmt.Errorf("unknown lifecycle mode %q", input.mode)
	}
}

func runLocate(input options) error {
	runRaw, runExit := readLookup(input.runPath, input.runLookupExit)
	artifactsRaw, artifactsExit := readLookup(input.artifactsPath, input.artifactsLookupExit)
	repository := os.DirFS(filepath.Dir(input.contract))
	metadata, receipt := selfimprovementtransport.ObserveArtifactLifecycle(
		repository, filepath.Base(input.contract), runRaw, artifactsRaw,
		selfimprovementtransport.ArtifactLifecycleInput{
			Selection: selfimprovementtransport.ArtifactSelectionInput{
				Repository: input.repository, ExpectedRunID: input.runID,
				ExpectedRunAttempt: input.runAttempt, ArtifactName: input.artifactName,
			},
			RunLookupExit: runExit, ArtifactsLookupExit: artifactsExit,
		})
	if err := writeReceipt(input.receipt, receipt); err != nil {
		return err
	}
	if metadata.ArtifactID > 0 {
		if err := writeJSON(input.output, metadata); err != nil {
			return err
		}
	}
	printReceipt(receipt)
	return nil
}

func runTransport(input options) error {
	raw, err := os.ReadFile(input.receipt)
	if err != nil {
		return fmt.Errorf("read lifecycle receipt: %w", err)
	}
	var receipt selfimprovementtransport.LifecycleReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return fmt.Errorf("decode lifecycle receipt: %w", err)
	}
	archiveRaw, downloadExit := readLookup(input.archive, input.downloadExit)
	receipt = selfimprovementtransport.CompleteArtifactLifecycle(receipt, archiveRaw, downloadExit)
	if err := writeReceipt(input.receipt, receipt); err != nil {
		return err
	}
	printReceipt(receipt)
	return nil
}

func readLookup(path string, exitCode int) ([]byte, int) {
	if exitCode != 0 {
		return nil, exitCode
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 1
	}
	return raw, 0
}

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
