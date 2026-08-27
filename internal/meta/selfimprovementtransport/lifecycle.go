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
	verifyLifecycleStep(&receipt, 2, digestJSON(metadata))
	closeLifecycle(&receipt)
	return metadata, receipt
}

func CompleteArtifactLifecycle(receipt LifecycleReceipt, archiveRaw []byte, downloadExit int) LifecycleReceipt {
	if len(receipt.Indicators) != lifecycleFixedStepTotal ||
		receipt.Indicators[0].Status != StatusVerified ||
		receipt.Indicators[1].Status != StatusVerified ||
		receipt.Indicators[2].Status != StatusVerified {
		closeLifecycle(&receipt)
		return receipt
	}
	if downloadExit != 0 || len(archiveRaw) == 0 {
		failLifecycleAt(&receipt, 3, "ARTIFACT_DOWNLOAD_FAILED", "", "", "")
		closeLifecycle(&receipt)
		return receipt
	}
	actual := digestBytes(archiveRaw)
	receipt.ActualArchiveDigest = actual
	verifyLifecycleStep(&receipt, 3, actual)
	if actual != receipt.ArtifactDigest {
		failLifecycleAt(&receipt, 4, "ARCHIVE_DIGEST_MISMATCH",
			receipt.ArtifactDigest, actual, "CONTRADICTION")
		closeLifecycle(&receipt)
		return receipt
	}
	verifyLifecycleStep(&receipt, 4, actual)
	closeLifecycle(&receipt)
	return receipt
}

func newLifecycleReceipt(contract ContractEvidence, input ArtifactSelectionInput) LifecycleReceipt {
	receipt := LifecycleReceipt{
		Schema: LifecycleReceiptSchema, MetricID: LifecycleMetricID,
		DenominatorID: LifecycleDenominatorID, Contract: contract,
		Repository: input.Repository, ExpectedRunID: input.ExpectedRunID,
		ExpectedRunAttempt: input.ExpectedRunAttempt, ArtifactName: input.ArtifactName,
		EnforcementEffect: LifecycleEffectNoEffect,
	}
	for index, definition := range lifecycleDefinitions {
		receipt.Indicators = append(receipt.Indicators, LifecycleIndicator{
			Ordinal: index + 1, MetricID: LifecycleMetricID, Class: definition.Class,
			ProofChoice: definition.ProofChoice,
			Coordinate: Coordinate{Stage: definition.Stage, Step: definition.Step},
			MetaOperation: definition.MetaOperation, Target: 1,
			Status: StatusUnknown, Reason: "LIFECYCLE_STEP_NOT_RUN",
		})
	}
	return receipt
}

func verifyLifecycleStep(receipt *LifecycleReceipt, index int, evidenceDigest string) {
	indicator := &receipt.Indicators[index]
	indicator.Status, indicator.Value = StatusVerified, 1
	indicator.Reason = lifecycleDefinitions[index].SuccessReason
	indicator.ObservationClass = "EVIDENCE"
	indicator.ExpectedDigest, indicator.ObservedDigest = "", ""
	indicator.EvidenceDigest = evidenceDigest
}

func failLifecycleAt(receipt *LifecycleReceipt, index int, reason, expected, observed, class string) {
	indicator := &receipt.Indicators[index]
	indicator.Status, indicator.Value, indicator.Reason = StatusUnknown, 0, reason
	indicator.ExpectedDigest, indicator.ObservedDigest = expected, observed
	indicator.ObservationClass, indicator.EvidenceDigest = class, ""
	for next := index + 1; next < len(receipt.Indicators); next++ {
		receipt.Indicators[next].Status = StatusUnknown
		receipt.Indicators[next].Value = 0
		receipt.Indicators[next].Reason = "UPSTREAM_ARTIFACT_STATE_UNKNOWN"
		receipt.Indicators[next].EvidenceDigest = ""
	}
}

func lifecycleSelectionReason(err error) string {
	coordinate := strings.SplitN(err.Error(), ":", 2)[0]
	parts := strings.Split(coordinate, "/")
	reason := parts[len(parts)-1]
	if strings.HasPrefix(reason, "SOURCE_RUN_") || reason == "SOURCE_WORKFLOW_PATH_UNKNOWN" {
		return "ARTIFACT_RUN_BINDING_MISMATCH"
	}
	return reason
}
