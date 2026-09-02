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
	artifact, err := resolveArtifact(artifacts, run, name)
	if err != nil {
		return artifactAPI{}, err
	}
	if err := validateArtifact(artifact); err != nil {
		return artifactAPI{}, err
	}
	return artifact, nil
}

func resolveArtifact(artifacts []artifactAPI, run workflowRunAPI,
	name string) (artifactAPI, error) {
	named := make([]artifactAPI, 0, 1)
	bound := make([]artifactAPI, 0, 1)
	for _, artifact := range artifacts {
		if artifact.Name != name {
			continue
		}
		named = append(named, artifact)
		if artifact.WorkflowRun.ID == run.ID && artifact.WorkflowRun.HeadSHA == run.HeadSHA {
			bound = append(bound, artifact)
		}
	}
	if len(named) == 0 {
		return artifactAPI{}, fmt.Errorf("LOCATE/select-immutable-artifact/ARTIFACT_NOT_FOUND")
	}
	if len(bound) == 0 {
		return artifactAPI{}, fmt.Errorf("LOCATE/resolve-artifact/ARTIFACT_RUN_BINDING_MISMATCH")
	}
	if len(bound) != 1 {
		return artifactAPI{}, fmt.Errorf(
			"LOCATE/select-immutable-artifact/ARTIFACT_SELECTION_AMBIGUOUS: got %d", len(bound))
	}
	return bound[0], nil
}

func validateArtifact(artifact artifactAPI) error {
	if artifact.Expired {
		return fmt.Errorf("LOCATE/validate-artifact-metadata/ARTIFACT_EXPIRED")
	}
	if artifact.ID <= 0 || artifact.SizeInBytes <= 0 || !validDigest(artifact.Digest) {
		return fmt.Errorf("LOCATE/validate-artifact-metadata/ARTIFACT_METADATA_INVALID")
	}
	return nil
}
