package selfimprovementtransport

import (
	"fmt"
	"strings"
)

func validateSelectionRun(run workflowRunAPI, input ArtifactSelectionInput) error {
	if run.ID != input.ExpectedRunID {
		return fmt.Errorf("LOCATE/verify-run-id/SOURCE_RUN_ID_MISMATCH: got %d want %d",
			run.ID, input.ExpectedRunID)
	}
	if run.RunAttempt != input.ExpectedRunAttempt {
		return fmt.Errorf("LOCATE/verify-run-attempt/SOURCE_RUN_ATTEMPT_MISMATCH: got %d want %d",
			run.RunAttempt, input.ExpectedRunAttempt)
	}
	if !validSHA(run.HeadSHA) {
		return fmt.Errorf("LOCATE/verify-run-head/SOURCE_RUN_HEAD_UNKNOWN")
	}
	if strings.TrimSpace(run.Path) == "" {
		return fmt.Errorf("LOCATE/verify-workflow-path/SOURCE_WORKFLOW_PATH_UNKNOWN")
	}
	return nil
}

func selectArtifact(artifacts []artifactAPI, run workflowRunAPI,
	name string) (artifactAPI, error) {
	matches := make([]artifactAPI, 0, 1)
	for _, artifact := range artifacts {
		if artifact.Name == name && !artifact.Expired &&
			artifact.WorkflowRun.ID == run.ID && artifact.WorkflowRun.HeadSHA == run.HeadSHA {
			matches = append(matches, artifact)
		}
	}
	if len(matches) == 0 {
		return artifactAPI{}, fmt.Errorf("LOCATE/select-immutable-artifact/ARTIFACT_NOT_FOUND")
	}
	if len(matches) != 1 {
		return artifactAPI{}, fmt.Errorf("LOCATE/select-immutable-artifact/ARTIFACT_SELECTION_AMBIGUOUS: got %d", len(matches))
	}
	artifact := matches[0]
	if artifact.ID <= 0 || artifact.SizeInBytes <= 0 || !validDigest(artifact.Digest) {
		return artifactAPI{}, fmt.Errorf("LOCATE/validate-immutable-artifact/ARTIFACT_IDENTITY_UNKNOWN")
	}
	return artifact, nil
}
