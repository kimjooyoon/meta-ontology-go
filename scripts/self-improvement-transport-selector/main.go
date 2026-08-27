package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func main() {
	mode := flag.String("mode", "locate", "locate metadata or seal transported archive")
	contract := flag.String("contract", "", "Gooo transport contract")
	runPath := flag.String("run", "", "GitHub workflow run API response")
	artifactsPath := flag.String("artifacts", "", "GitHub artifact list API response")
	repository := flag.String("repository", "", "owner/repository")
	runID := flag.Int64("expected-run-id", 0, "expected producer workflow run id")
	runAttempt := flag.Int("expected-run-attempt", 0, "expected producer run attempt")
	artifactName := flag.String("artifact-name", selfimprovementtransport.ArtifactName, "artifact lookup label")
	output := flag.String("output", "", "selected transport metadata output")
	receipt := flag.String("receipt", "", "artifact lifecycle receipt")
	archive := flag.String("archive", "", "downloaded artifact archive")
	runLookupExit := flag.Int("run-lookup-exit", 0, "workflow run metadata lookup exit status")
	artifactsLookupExit := flag.Int("artifacts-lookup-exit", 0, "artifact list lookup exit status")
	downloadExit := flag.Int("download-exit", 0, "artifact archive download exit status")
	flag.Parse()
	if err := run(options{
		mode: *mode, contract: *contract, runPath: *runPath, artifactsPath: *artifactsPath,
		repository: *repository, runID: *runID, runAttempt: *runAttempt,
		artifactName: *artifactName, output: *output, receipt: *receipt, archive: *archive,
		runLookupExit: *runLookupExit, artifactsLookupExit: *artifactsLookupExit,
		downloadExit: *downloadExit,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
