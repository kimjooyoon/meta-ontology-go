package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func run(runPath, artifactsPath, repository string, runID int64, runAttempt int,
	artifactName, output string) error {
	runRaw, err := os.ReadFile(runPath)
	if err != nil {
		return fmt.Errorf("LOCATE/read-run/RUN_RESPONSE_UNAVAILABLE: %w", err)
	}
	artifactsRaw, err := os.ReadFile(artifactsPath)
	if err != nil {
		return fmt.Errorf("LOCATE/read-artifacts/ARTIFACT_RESPONSE_UNAVAILABLE: %w", err)
	}
	metadata, err := selfimprovementtransport.SelectTransportMetadata(runRaw, artifactsRaw,
		selfimprovementtransport.ArtifactSelectionInput{Repository: repository,
			ExpectedRunID: runID, ExpectedRunAttempt: runAttempt, ArtifactName: artifactName})
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("LOCATE/encode-selection/SELECTION_RECEIPT_INVALID: %w", err)
	}
	if err := os.WriteFile(output, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("LOCATE/write-selection/SELECTION_RECEIPT_UNAVAILABLE: %w", err)
	}
	fmt.Printf("transport selection: run=%d attempt=%d artifact=%d\n",
		metadata.ProducerRunID, metadata.ProducerRunAttempt, metadata.ArtifactID)
	return nil
}
