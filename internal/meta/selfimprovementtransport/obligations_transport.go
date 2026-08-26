package selfimprovementtransport

func logicalDigestObligation(input evaluationInput) Obligation {
	actual := digestBytes(input.observationRaw)
	return knownObligation("logical-subject-digest", "REGRESSION", "PRODUCE", "hash-logical-subject",
		input.producerErr == nil && input.producer.Subject.Name == "first.json" &&
			input.producer.Subject.Bytes == len(input.observationRaw) &&
			input.producer.Subject.Digest == actual && validDigest(input.producer.Subject.Digest),
		"LOGICAL_SUBJECT_DIGEST_VERIFIED", "LOGICAL_SUBJECT_DIGEST_MISMATCH",
		struct{ Expected, Actual string }{input.producer.Subject.Digest, actual})
}

func immutableLocatorObligation(input evaluationInput) Obligation {
	passed := input.metadataErr == nil && input.metadata.Schema == MetadataSchema &&
		input.metadata.Repository == input.expectedRepository && input.metadata.ArtifactID > 0 &&
		input.metadata.ArtifactName == ArtifactName && input.metadata.ArtifactSizeBytes > 0 &&
		input.metadata.ProducerRunID == input.expectedRunID && input.metadata.ProducerRunAttempt > 0
	return knownObligation("immutable-artifact-locator", "FOUNDATION", "TRANSPORT", "locate-immutable-artifact",
		passed, "IMMUTABLE_ARTIFACT_LOCATED", "IMMUTABLE_ARTIFACT_LOCATOR_MISMATCH", input.metadata)
}

func archiveDigestObligation(input evaluationInput) Obligation {
	return knownObligation("artifact-archive-digest", "REGRESSION", "TRANSPORT", "verify-archive-digest",
		input.metadataErr == nil && validDigest(input.metadata.ArtifactDigest) &&
			validDigest(input.actualArchiveDigest) && input.metadata.ArtifactDigest == input.actualArchiveDigest,
		"ARCHIVE_DIGEST_REPLAYED", "ARCHIVE_DIGEST_MISMATCH",
		struct{ Expected, Actual string }{input.metadata.ArtifactDigest, input.actualArchiveDigest})
}

func consumerReplayObligation(input evaluationInput) Obligation {
	passed := input.contractErr == nil && input.producerErr == nil && input.sourceErr == nil &&
		ValidateProducer(input.producer, input.source, input.observationRaw, input.contract) == nil &&
		input.producer.Digest == producerDigest(input.producer)
	evidence := struct{ Producer, Payload, Contract string }{
		input.producer.Digest, digestBytes(input.observationRaw), input.contract.CanonicalDigest,
	}
	return knownObligation("consumer-digest-replay", "REGRESSION", "CONSUME", "recompute-subject-digest",
		passed, "CONSUMER_DIGEST_REPLAYED", "CONSUMER_DIGEST_REPLAY_MISMATCH", evidence)
}
