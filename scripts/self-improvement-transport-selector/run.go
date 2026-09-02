package main

import (
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
