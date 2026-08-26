package selfimprovementtransport

import (
	"encoding/json"
	"io/fs"
)

var obligationOrder = []string{
	"source-repository-commit",
	"producer-checkout-binding",
	"producer-run-identity",
	"logical-subject-digest",
	"immutable-artifact-locator",
	"artifact-archive-digest",
	"consumer-digest-replay",
	attestationObligation,
}

func Evaluate(repository fs.FS, contractPath, expectedRepository string, expectedRunID int64,
	observationRaw, producerRaw, metadataRaw []byte, actualArchiveDigest string) Report {
	contract, contractErr := CompileContract(repository, contractPath)
	var source observationHeader
	var producer ProducerReceipt
	var metadata TransportMetadata
	sourceErr := json.Unmarshal(observationRaw, &source)
	producerErr := json.Unmarshal(producerRaw, &producer)
	metadataErr := json.Unmarshal(metadataRaw, &metadata)
	report := Report{
		Schema: ReportSchema, MetricID: MetricID, Contract: contract,
		SubjectSHA: source.SubjectSHA, OrchestrationHeadSHA: metadata.OrchestrationHeadSHA,
		SourceObservationDigest: digestBytes(observationRaw), ActualArchiveDigest: actualArchiveDigest,
		Producer: producer, Transport: metadata,
		NotClaimed: []string{"artifact-name-proves-subject", "archive-digest-authenticates-producer", "whole-language-transport-complete"},
	}
	expectedURI := "https://github.com/" + expectedRepository
	report.Obligations = append(report.Obligations,
		knownObligation("source-repository-commit", "FOUNDATION", "SOURCE", "validate-repository-commit",
			sourceErr == nil && producerErr == nil && source.Schema == ObservationSchema &&
				validSHA(source.SubjectSHA) && producer.SubjectSHA == source.SubjectSHA && producer.RepositoryURI == expectedURI,
			"SOURCE_REPOSITORY_COMMIT_VERIFIED", "SOURCE_REPOSITORY_COMMIT_MISMATCH",
			struct{ Repository, Subject string }{producer.RepositoryURI, source.SubjectSHA}),
		knownObligation("producer-checkout-binding", "COHERENCE", "PRODUCE", "verify-checkout-head",
			producerErr == nil && validSHA(producer.CheckoutSHA) && producer.CheckoutSHA == source.SubjectSHA,
			"PRODUCER_CHECKOUT_BOUND", "PRODUCER_CHECKOUT_MISMATCH",
			struct{ Checkout, Subject string }{producer.CheckoutSHA, source.SubjectSHA}),
		knownObligation("producer-run-identity", "FOUNDATION", "PRODUCE", "bind-run-identity",
			producerErr == nil && metadataErr == nil && expectedRunID > 0 && producer.RunID == expectedRunID &&
				metadata.ProducerRunID == expectedRunID && producer.RunAttempt == metadata.ProducerRunAttempt &&
				producer.WorkflowRef != "" && validSHA(producer.WorkflowSHA) && producer.Job != "" &&
				metadata.WorkflowPath != "" && validSHA(metadata.OrchestrationHeadSHA),
			"PRODUCER_RUN_IDENTITY_BOUND", "PRODUCER_RUN_IDENTITY_MISMATCH",
			struct{ RunID int64; Attempt int; Workflow string }{metadata.ProducerRunID, metadata.ProducerRunAttempt, metadata.WorkflowPath}),
		knownObligation("logical-subject-digest", "REGRESSION", "PRODUCE", "hash-logical-subject",
			producerErr == nil && producer.Subject.Name == "first.json" && producer.Subject.Bytes == len(observationRaw) &&
				producer.Subject.Digest == digestBytes(observationRaw) && validDigest(producer.Subject.Digest),
			"LOGICAL_SUBJECT_DIGEST_VERIFIED", "LOGICAL_SUBJECT_DIGEST_MISMATCH",
			struct{ Expected, Actual string }{producer.Subject.Digest, digestBytes(observationRaw)}),
		knownObligation("immutable-artifact-locator", "FOUNDATION", "TRANSPORT", "locate-immutable-artifact",
			metadataErr == nil && metadata.Schema == MetadataSchema && metadata.Repository == expectedRepository &&
				metadata.ArtifactID > 0 && metadata.ArtifactName == ArtifactName && metadata.ArtifactSizeBytes > 0 &&
				metadata.ProducerRunID == expectedRunID && metadata.ProducerRunAttempt > 0,
			"IMMUTABLE_ARTIFACT_LOCATED", "IMMUTABLE_ARTIFACT_LOCATOR_MISMATCH", metadata),
		knownObligation("artifact-archive-digest", "REGRESSION", "TRANSPORT", "verify-archive-digest",
			metadataErr == nil && validDigest(metadata.ArtifactDigest) && validDigest(actualArchiveDigest) &&
				metadata.ArtifactDigest == actualArchiveDigest,
			"ARCHIVE_DIGEST_REPLAYED", "ARCHIVE_DIGEST_MISMATCH",
			struct{ Expected, Actual string }{metadata.ArtifactDigest, actualArchiveDigest}),
		knownObligation("consumer-digest-replay", "REGRESSION", "CONSUME", "recompute-subject-digest",
			contractErr == nil && producerErr == nil && sourceErr == nil &&
				ValidateProducer(producer, source, observationRaw, contract) == nil && producer.Digest == producerDigest(producer),
			"CONSUMER_DIGEST_REPLAYED", "CONSUMER_DIGEST_REPLAY_MISMATCH",
			struct{ Producer, Payload, Contract string }{producer.Digest, digestBytes(observationRaw), contract.CanonicalDigest}),
	)
	report.Obligations = append(report.Obligations, attestationObligationFor(metadata.Attestation))
	reduce(&report)
	report.Digest = reportDigest(report)
	return report
}

func knownObligation(id, route, stage, step string, passed bool, successReason, failureReason string, evidence any) Obligation {
	status, reason := StatusFalse, failureReason
	if passed {
		status, reason = StatusVerified, successReason
	}
	return Obligation{ID: id, ProofRoute: route, Coordinate: Coordinate{Stage: stage, Step: step},
		Status: status, Reason: reason, EvidenceDigest: digestJSON(evidence)}
}

func attestationObligationFor(attestation Attestation) Obligation {
	obligation := Obligation{ID: attestationObligation, ProofRoute: "COHERENCE",
		Coordinate: Coordinate{Stage: "ATTEST", Step: "verify-producer-identity"}}
	switch attestation.Status {
	case "VERIFIED":
		if validDigest(attestation.Digest) && attestation.ProducerIdentity != "" {
			obligation.Status, obligation.Reason = StatusVerified, "PRODUCER_ATTESTATION_VERIFIED"
			obligation.EvidenceDigest = digestJSON(attestation)
			return obligation
		}
		obligation.Status, obligation.Reason = StatusFalse, "PRODUCER_ATTESTATION_INVALID"
		obligation.EvidenceDigest = digestJSON(attestation)
	case "", "UNKNOWN":
		obligation.Status, obligation.Reason = StatusUnknown, ReasonAttestation
	default:
		obligation.Status, obligation.Reason = StatusFalse, "PRODUCER_ATTESTATION_REJECTED"
		obligation.EvidenceDigest = digestJSON(attestation)
	}
	return obligation
}

func reduce(report *Report) {
	report.Metrics.FixedObligationTotal = fixedObligationTotal
	var firstFalse, firstUnknown *Obligation
	for index := range report.Obligations {
		obligation := &report.Obligations[index]
		switch obligation.Status {
		case StatusVerified:
			report.Metrics.VerifiedTotal++
		case StatusUnknown:
			report.Metrics.UnknownTotal++
			report.OpenObligationIDs = append(report.OpenObligationIDs, obligation.ID)
			if firstUnknown == nil {
				firstUnknown = obligation
			}
		default:
			report.Metrics.FalseTotal++
			report.OpenObligationIDs = append(report.OpenObligationIDs, obligation.ID)
			if firstFalse == nil {
				firstFalse = obligation
			}
		}
	}
	report.Metrics.OpenTotal = report.Metrics.UnknownTotal + report.Metrics.FalseTotal
	report.Metrics.CoverageBasisPoints = report.Metrics.VerifiedTotal * 10000 / fixedObligationTotal
	switch {
	case firstFalse != nil:
		report.Decision, report.Resolution, report.Reason = DecisionFailClosed, ResolutionLower, ReasonKnownMismatch
		report.Coordinate = firstFalse.Coordinate
	case firstUnknown != nil:
		report.Decision, report.Resolution, report.Reason = DecisionObserved, ResolutionLower, firstUnknown.Reason
		report.Coordinate = firstUnknown.Coordinate
	default:
		report.Decision, report.Resolution, report.Reason = DecisionPass, ResolutionExact, ReasonComplete
		report.Coordinate = Coordinate{Stage: "REDUCE", Step: "close-eht8"}
	}
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestJSON(report)
}
