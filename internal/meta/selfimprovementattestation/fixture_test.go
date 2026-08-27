package selfimprovementattestation

import "strings"

func validRequest() Request {
	commit := strings.Repeat("1", 40)
	workflowSHA := strings.Repeat("2", 40)
	archiveDigest := "sha256:" + strings.Repeat("a", 64)
	observationDigest := "sha256:" + strings.Repeat("b", 64)
	contract := Contract{ContractID: "gooo.contract.self-improvement.exact-head-transport.v1", Path: "transport.gooo", Package: "selfimprovementtransport", Namespace: "selfimprovementtransport", ObligationTotal: 8}
	producer := Producer{
		Schema: "gooo/self-improvement-transport-producer/v1", Contract: contract,
		RepositoryURI: "https://github.com/owner/repo", SubjectSHA: commit, CheckoutSHA: commit,
		WorkflowRef: "owner/repo/.github/workflows/self-improvement-language-observation.yml@refs/heads/dev",
		WorkflowSHA: workflowSHA, RunID: 12, RunAttempt: 1, Job: "observation",
		ArtifactName: "self-improvement-language-observation",
		LogicalSubject: LogicalSubject{Name: "first.json", Digest: observationDigest, Bytes: 10},
		Decision: "BOUND", Resolution: "EXACT", Reason: "PRODUCER_SUBJECT_BOUND", Digest: "sha256:producer",
	}
	obligations := fixtureObligations()
	prior := TransportReceipt{
		Schema: transportSchema, MetricID: metricID, Contract: contract, SubjectSHA: commit,
		OrchestrationHeadSHA: commit, SourceObservationDigest: observationDigest,
		ActualArchiveDigest: archiveDigest, Decision: "OBSERVED", Resolution: "LOWER_RESOLUTION",
		Reason: "PRODUCER_ATTESTATION_UNKNOWN", Coordinate: Coordinate{Stage: "ATTEST", Step: "verify-producer-identity"},
		Producer: producer, Transport: Transport{Repository: "owner/repo", ProducerRunID: 12, ProducerRunAttempt: 1,
			OrchestrationHeadSHA: commit, WorkflowPath: ".github/workflows/self-improvement-language-observation.yml",
			ArtifactID: 34, ArtifactName: producer.ArtifactName, ArtifactDigest: archiveDigest},
		Obligations: obligations, OpenObligationIDs: []string{attestationID}, Metrics: Metrics{8, 7, 1, 0, 1, 8750, 0}, Digest: "sha256:prior",
	}
	result := fixtureVerification(producer, archiveDigest)
	return Request{TransportReceipt: prior, ArchiveDigest: archiveDigest, ArchiveProducer: producer,
		ArchiveObservationDigest: observationDigest, Verification: []VerificationItem{{VerificationResult: result}},
		VerifierExitCode: 0, VerifierVersion: "gh version fixture"}
}

func fixtureObligations() []Obligation {
	ids := []string{"source-repository-commit", "producer-checkout-binding", "producer-run-identity", "logical-subject-digest", "immutable-artifact-locator", "artifact-archive-digest", "consumer-digest-replay"}
	result := make([]Obligation, 0, 8)
	for _, id := range ids {
		result = append(result, Obligation{ID: id, ProofRoute: "FOUNDATION", Status: "VERIFIED", Reason: "VERIFIED", EvidenceDigest: "sha256:evidence"})
	}
	return append(result, Obligation{ID: attestationID, ProofRoute: "COHERENCE", Coordinate: Coordinate{Stage: "ATTEST", Step: "verify-producer-identity"}, Status: "UNKNOWN", Reason: "PRODUCER_ATTESTATION_UNKNOWN"})
}
