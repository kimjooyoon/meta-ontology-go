package selfimprovementtransport

import (
	"encoding/json"
	"io/fs"
	"strings"
)

func ObserveArtifactLifecycle(repository fs.FS, contractPath string, runRaw, artifactsRaw []byte,
	input ArtifactLifecycleInput) (TransportMetadata, LifecycleReceipt) {
	contract, contractErr := CompileContract(repository, contractPath)
	receipt := newLifecycleReceipt(contract, input.Selection)
	if contractErr != nil {
		failLifecycleAt(&receipt, 0, "LIFECYCLE_CONTRACT_INVALID", "", "", "")
		closeLifecycle(&receipt)
		return TransportMetadata{}, receipt
	}
	if input.Selection.ExpectedRunID <= 0 || input.Selection.ExpectedRunAttempt <= 0 ||
		strings.TrimSpace(input.Selection.Repository) == "" ||
		strings.TrimSpace(input.Selection.ArtifactName) == "" {
		failLifecycleAt(&receipt, 0, "ARTIFACT_LIFECYCLE_INPUT_INVALID", "", "", "")
		closeLifecycle(&receipt)
		return TransportMetadata{}, receipt
	}
	if input.RunLookupExit != 0 || input.ArtifactsLookupExit != 0 {
		failLifecycleAt(&receipt, 0, "ARTIFACT_METADATA_LOOKUP_FAILED", "", "", "")
		closeLifecycle(&receipt)
		return TransportMetadata{}, receipt
	}
	var run workflowRunAPI
	var list artifactListAPI
	if json.Unmarshal(runRaw, &run) != nil || json.Unmarshal(artifactsRaw, &list) != nil || list.Artifacts == nil {
		failLifecycleAt(&receipt, 0, "ARTIFACT_METADATA_RESPONSE_INVALID", "", "", "")
		closeLifecycle(&receipt)
		return TransportMetadata{}, receipt
	}
	verifyLifecycleStep(&receipt, 0, digestJSON([]string{digestBytes(runRaw), digestBytes(artifactsRaw)}))
	receipt.OrchestrationHeadSHA = run.HeadSHA
	receipt.WorkflowPath = run.Path
	receipt.ArtifactInstanceCount, receipt.ArtifactTypeCount = artifactStats(list.Artifacts, input.Selection.ArtifactName)
	if err := validateSelectionRun(run, input.Selection); err != nil {
		failLifecycleAt(&receipt, 1, lifecycleSelectionReason(err), "", "", "")
		closeLifecycle(&receipt)
		return TransportMetadata{}, receipt
	}
	artifact, err := resolveArtifact(list.Artifacts, run, input.Selection.ArtifactName)
	if err != nil {
		failLifecycleAt(&receipt, 1, lifecycleSelectionReason(err), "", "", "")
		closeLifecycle(&receipt)
		return TransportMetadata{}, receipt
	}
	verifyLifecycleStep(&receipt, 1, digestJSON([]any{artifact.ID, artifact.Name, run.ID, run.HeadSHA}))
	if err := validateArtifact(artifact); err != nil {
		failLifecycleAt(&receipt, 2, lifecycleSelectionReason(err), artifact.Digest, "", "")
		closeLifecycle(&receipt)
		return TransportMetadata{}, receipt
	}
	metadata := TransportMetadata{
		Schema: MetadataSchema, Repository: input.Selection.Repository,
		ProducerRunID: run.ID, ProducerRunAttempt: run.RunAttempt,
		OrchestrationHeadSHA: run.HeadSHA, WorkflowPath: run.Path,
		ArtifactID: artifact.ID, ArtifactName: artifact.Name,
		ArtifactDigest: artifact.Digest, ArtifactSizeBytes: artifact.SizeInBytes,
	}
	receipt.ArtifactID = artifact.ID
	receipt.ArtifactName = artifact.Name
	receipt.ArtifactDigest = artifact.Digest
	receipt.ArtifactSizeBytes = artifact.SizeInBytes
	verifyLifecycleStep(&receipt, 2, digestJSON(metadata))
	closeLifecycle(&receipt)
	return metadata, receipt
}
