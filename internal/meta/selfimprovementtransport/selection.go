package selfimprovementtransport

import (
	"encoding/json"
	"fmt"
	"strings"
)

func SelectTransportMetadata(runRaw, artifactsRaw []byte,
	input ArtifactSelectionInput) (TransportMetadata, error) {
	if input.ExpectedRunID <= 0 || input.ExpectedRunAttempt <= 0 ||
		strings.TrimSpace(input.Repository) == "" || strings.TrimSpace(input.ArtifactName) == "" {
		return TransportMetadata{}, fmt.Errorf("LOCATE/validate-request/SELECTION_INPUT_UNKNOWN")
	}
	var run workflowRunAPI
	if err := json.Unmarshal(runRaw, &run); err != nil {
		return TransportMetadata{}, fmt.Errorf("LOCATE/decode-run/RUN_RESPONSE_INVALID: %w", err)
	}
	if err := validateSelectionRun(run, input); err != nil {
		return TransportMetadata{}, err
	}
	var list artifactListAPI
	if err := json.Unmarshal(artifactsRaw, &list); err != nil {
		return TransportMetadata{}, fmt.Errorf("LOCATE/decode-artifacts/ARTIFACT_RESPONSE_INVALID: %w", err)
	}
	artifact, err := selectArtifact(list.Artifacts, run, input.ArtifactName)
	if err != nil {
		return TransportMetadata{}, err
	}
	instances, types := artifactStats(list.Artifacts, input.ArtifactName)
	return TransportMetadata{
		Schema: MetadataSchema, Repository: input.Repository,
		ProducerRunID: run.ID, ProducerRunAttempt: run.RunAttempt,
		OrchestrationHeadSHA: run.HeadSHA, WorkflowPath: run.Path,
		ArtifactID: artifact.ID, ArtifactName: artifact.Name,
		ArtifactDigest: artifact.Digest, ArtifactSizeBytes: artifact.SizeInBytes,
		ArtifactInstanceCount: instances, ArtifactTypeCount: types,
	}, nil
}
