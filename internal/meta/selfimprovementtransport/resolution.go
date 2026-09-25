package selfimprovementtransport

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
)

func BindTransportMetadata(receipt LifecycleReceipt, archiveRaw []byte) (TransportMetadata, error) {
	metadata := TransportMetadata{
		Schema: MetadataSchema, Repository: receipt.Repository,
		ProducerRunID: receipt.ExpectedRunID, ProducerRunAttempt: receipt.ExpectedRunAttempt,
		OrchestrationHeadSHA: receipt.OrchestrationHeadSHA, WorkflowPath: receipt.WorkflowPath,
		ArtifactID: receipt.ArtifactID, ArtifactName: receipt.ArtifactName,
		ArtifactDigest: receipt.ArtifactDigest, ArtifactSizeBytes: receipt.ArtifactSizeBytes,
		ArtifactInstanceCount: receipt.ArtifactInstanceCount, ArtifactTypeCount: receipt.ArtifactTypeCount,
	}
	files, err := readArchiveFiles(archiveRaw)
	if err != nil {
		return metadata, err
	}
	producerFiles := files["producer.json"]
	metadata.ProducerDeclarationCount = len(producerFiles)
	payloadFiles := files["first.json"]
	metadata.ProducerPayloadCount = len(payloadFiles)
	if len(producerFiles) != 1 {
		return metadata, nil
	}
	producerRaw := producerFiles[0]
	metadata.ProducerDeclarationDigest = digestBytes(producerRaw)
	var producer ProducerReceipt
	if err := json.Unmarshal(producerRaw, &producer); err != nil {
		return metadata, nil
	}
	metadata.ProducerSubjectSHA = producer.SubjectSHA
	if len(payloadFiles) == 1 {
		metadata.ProducerPayloadName = "first.json"
		metadata.ProducerPayloadDigest = digestBytes(payloadFiles[0])
		metadata.ProducerPayloadBytes = len(payloadFiles[0])
	}
	return metadata, nil
}

func readArchiveFiles(archiveRaw []byte) (map[string][][]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archiveRaw), int64(len(archiveRaw)))
	if err != nil {
		return nil, fmt.Errorf("read producer transport archive: %w", err)
	}
	files := map[string][][]byte{}
	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		name := path.Base(entry.Name)
		if name != "producer.json" && name != "first.json" {
			continue
		}
		stream, err := entry.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s in producer transport archive: %w", name, err)
		}
		data, readErr := io.ReadAll(io.LimitReader(stream, 8<<20))
		closeErr := stream.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read %s in producer transport archive: %w", name, readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close %s in producer transport archive: %w", name, closeErr)
		}
		files[name] = append(files[name], data)
	}
	return files, nil
}

func evaluateProvenance(input evaluationInput) ProvenanceResolution {
	resolution := ProvenanceResolution{
		State: ResolutionUnknown, Stage: "CONSUME", Step: "resolve-producer-declaration",
		Reason: "PRODUCER_DECLARATION_MISSING", ArtifactInstances: input.metadata.ArtifactInstanceCount,
		ArtifactTypes: input.metadata.ArtifactTypeCount,
	}
	if input.producerErr != nil && len(input.producerRaw) > 0 {
		return refutedResolution(resolution, "PRODUCER_DECLARATION_INVALID")
	}
	if input.metadataErr != nil || input.metadata.ProducerDeclarationCount == 0 || len(input.producerRaw) == 0 {
		return unknownResolution(resolution, "CONSUME", "resolve-producer-declaration", "PRODUCER_DECLARATION_MISSING", "DIRECT_MISSING", []string{"producer-declaration", "transport-index"})
	}
	if input.metadata.ProducerDeclarationCount != 1 || input.metadata.ArtifactInstanceCount > 1 {
		return refutedResolution(resolution, "DUPLICATE_PRODUCER_DECLARATION")
	}
	if input.metadata.ProducerPayloadCount > 1 {
		return refutedResolution(resolution, "DUPLICATE_PRODUCER_PAYLOAD")
	}
	if input.metadata.ArtifactTypeCount > 1 {
		return refutedResolution(resolution, "DUPLICATE_ARTIFACT_TYPE")
	}
	var producer ProducerReceipt
	if err := json.Unmarshal(input.producerRaw, &producer); err != nil {
		return refutedResolution(resolution, "PRODUCER_DECLARATION_INVALID")
	}
	if input.metadata.ArtifactInstanceCount == 0 || input.metadata.ArtifactTypeCount == 0 || input.metadata.ArtifactID <= 0 || input.metadata.ArtifactName == "" || input.metadata.ArtifactDigest == "" {
		return unknownResolution(resolution, "LOCATE", "read-artifact-metadata", "ARTIFACT_INDEX_MISSING", "DIRECT_MISSING", []string{"artifact-index", "artifact-id", "artifact-digest"})
	}
	if reason := producerIdentityContradiction(input, producer); reason != "" {
		return refutedResolution(resolution, reason)
	}
	if input.metadata.ProducerPayloadCount == 0 || input.sourceErr != nil || len(input.observationRaw) == 0 {
		return unknownResolution(resolution, "CONSUME", "resolve-payload", "PRODUCER_PAYLOAD_MISSING", "DIRECT_MISSING", []string{"payload", "producer-declaration"})
	}
	if reason := producerPayloadContradiction(input, producer); reason != "" {
		return refutedResolution(resolution, reason)
	}
	if input.actualArchiveDigest == "" {
		return unknownResolution(resolution, "TRANSPORT", "verify-archive-digest", "ARCHIVE_DIGEST_MISSING", "DIRECT_MISSING", []string{"artifact-archive", "artifact-digest"})
	}
	if !validDigest(input.actualArchiveDigest) || input.metadata.ArtifactDigest != input.actualArchiveDigest {
		return refutedResolution(resolution, "PRODUCER_ARTIFACT_DIGEST_MISMATCH")
	}
	resolution.State = ResolutionClosed
	resolution.Stage = "CONSUME"
	resolution.Step = "resolve-producer-subject"
	resolution.Reason = "EXACT_PRODUCER_SUBJECT_PAYLOAD_MATCH"
	resolution.ProducerDeclarationDigest = digestBytes(input.producerRaw)
	resolution.ProducerSubjectSHA = producer.SubjectSHA
	resolution.ProducerPayloadDigest = producer.Subject.Digest
	resolution.ProducerPayloadBytes = producer.Subject.Bytes
	return resolution
}

func producerIdentityContradiction(input evaluationInput, producer ProducerReceipt) string {
	if input.metadata.Schema != MetadataSchema || input.metadata.Repository != input.expectedRepository || producer.RepositoryURI != "https://github.com/"+input.expectedRepository {
		return "PRODUCER_REPOSITORY_WORKFLOW_MISMATCH"
	}
	if producer.ArtifactName != input.metadata.ArtifactName || input.metadata.ArtifactName == "" {
		return "PRODUCER_ARTIFACT_NAME_MISMATCH"
	}
	if producer.RunID != input.expectedRunID || producer.RunAttempt != input.metadata.ProducerRunAttempt || producer.RunAttempt <= 0 || input.metadata.ProducerRunID != input.expectedRunID {
		return "PRODUCER_RUN_ATTEMPT_MISMATCH"
	}
	if input.metadata.OrchestrationHeadSHA == "" || producer.SubjectSHA != input.metadata.OrchestrationHeadSHA {
		return "PRODUCER_SUBJECT_RUN_MISMATCH"
	}
	if input.metadata.WorkflowPath == "" || workflowPath(producer.WorkflowRef) != input.metadata.WorkflowPath {
		return "PRODUCER_REPOSITORY_WORKFLOW_MISMATCH"
	}
	if input.metadata.ProducerDeclarationDigest == "" || input.metadata.ProducerDeclarationDigest != digestBytes(input.producerRaw) {
		return "PRODUCER_DECLARATION_DIGEST_MISMATCH"
	}
	return ""
}

func producerPayloadContradiction(input evaluationInput, producer ProducerReceipt) string {
	if producer.Subject.Name != "first.json" || producer.Subject.Bytes != len(input.observationRaw) || producer.Subject.Digest != digestBytes(input.observationRaw) {
		return "PRODUCER_PAYLOAD_DIGEST_MISMATCH"
	}
	if input.source.Schema != ObservationSchema || input.source.SubjectSHA != producer.SubjectSHA {
		return "PRODUCER_PAYLOAD_SUBJECT_MISMATCH"
	}
	if input.metadata.ProducerSubjectSHA == "" || input.metadata.ProducerSubjectSHA != producer.SubjectSHA || input.metadata.ProducerPayloadName != producer.Subject.Name || input.metadata.ProducerPayloadDigest != producer.Subject.Digest || input.metadata.ProducerPayloadBytes != producer.Subject.Bytes {
		return "PRODUCER_PAYLOAD_SUBJECT_MISMATCH"
	}
	return ""
}

func workflowPath(value string) string {
	before, _, _ := strings.Cut(value, "@")
	if index := strings.Index(before, "/.github/workflows/"); index >= 0 {
		return before[index+1:]
	}
	return before
}

func unknownResolution(resolution ProvenanceResolution, stage, step, reason, class string, blockedBy []string) ProvenanceResolution {
	resolution.State, resolution.Stage, resolution.Step, resolution.Reason = ResolutionUnknown, stage, step, reason
	resolution.Unknown = &CausalUnknown{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: nextResolutionOperation(stage, step), BlockedBy: blockedBy}
	return resolution
}

func refutedResolution(resolution ProvenanceResolution, reason string) ProvenanceResolution {
	resolution.State, resolution.Reason, resolution.Unknown = ResolutionRefuted, reason, nil
	resolution.FallbackAttempted = strings.Contains(reason, "FALLBACK")
	resolution.FallbackAccepted = false
	return resolution
}

func nextResolutionOperation(stage, step string) string {
	if stage == "LOCATE" {
		return "retry_exact_producer_artifact_lookup"
	}
	if step == "resolve-payload" {
		return "restore_exact_producer_payload"
	}
	return "restore_exact_producer_declaration"
}

func resolutionMetrics(policy ResolutionPolicy, resolution ProvenanceResolution) ResolutionMetrics {
	metrics := ResolutionMetrics{
		CaseDenominator: policy.CaseDenominator, ClosedCases: policy.ClosedCases,
		UnknownCases: policy.UnknownCases, RefutedCases: policy.RefutedCases,
		ProvenanceStateBefore: ResolutionUnknown, ProvenanceStateAfter: ResolutionClosed,
		ArtifactInstancesBefore: 1, ArtifactInstancesAfter: 1,
		ArtifactTypesBefore: 1, ArtifactTypesAfter: 1,
	}
	for key, value := range policy.Metrics {
		switch key {
		case "active_root_before":
			metrics.ActiveRootBefore = value
		case "active_root_after":
			metrics.ActiveRootAfter = value
		case "exact_resolutions_before":
			metrics.ExactResolutionsBefore = value
		case "exact_resolutions_after":
			metrics.ExactResolutionsAfter = value
		case "unknown_six_field_before":
			metrics.UnknownSixFieldBefore = value
		case "unknown_six_field_after":
			metrics.UnknownSixFieldAfter = value
		case "refuted_contradictions_before":
			metrics.RefutedContradictionsBefore = value
		case "refuted_contradictions_after":
			metrics.RefutedContradictionsAfter = value
		case "fallback_accepted_before":
			metrics.FallbackAcceptedBefore = value
		case "fallback_accepted_after":
			metrics.FallbackAcceptedAfter = value
		case "artifact_instances_before":
			metrics.ArtifactInstancesBefore = value
		case "artifact_instances_after":
			metrics.ArtifactInstancesAfter = value
		case "artifact_types_before":
			metrics.ArtifactTypesBefore = value
		case "artifact_types_after":
			metrics.ArtifactTypesAfter = value
		case "independent_replay_comparisons_before":
			metrics.IndependentReplayComparisonsBefore = value
		case "independent_replay_comparisons_after":
			metrics.IndependentReplayComparisonsAfter = value
		}
	}
	switch resolution.State {
	case ResolutionClosed:
		metrics.CurrentExact = 1
	case ResolutionUnknown:
		metrics.CurrentUnknownSixField = boolInt(resolution.Unknown != nil)
	case ResolutionRefuted:
		metrics.CurrentRefuted = 1
	}
	if resolution.FallbackAccepted {
		metrics.FallbackAccepted = 1
	}
	return metrics
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
