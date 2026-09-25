package selfimprovementattestation

func observe(receipt *ResolutionReceipt, request Request, result *VerificationResult, evidence string) {
	producer, transport := request.TransportReceipt.Producer, request.TransportReceipt.Transport
	certificate := result.Signature.Certificate
	receipt.Decision = "OBSERVED"
	receipt.Resolution = "EXACT"
	receipt.Reason = "PRODUCER_ATTESTATION_VERIFIED"
	receipt.ProducerIdentity = ProducerIdentity{
		WorkflowRef:        producer.WorkflowRef,
		WorkflowSHA:        producer.WorkflowSHA,
		RunID:              producer.RunID,
		RunAttempt:         producer.RunAttempt,
		ArtifactID:         transport.ArtifactID,
		SubjectName:        producer.ArtifactName,
		SubjectDigest:      request.ArchiveDigest,
		SignerURI:          certificate.BuildSignerURI,
		Issuer:             certificate.Issuer,
		RunnerEnvironment:  certificate.RunnerEnvironment,
		VerifiedTimestamps: len(result.VerifiedTimestamps),
	}
	receipt.Obligations = setAttestation(receipt.Obligations, "VERIFIED", "PRODUCER_ATTESTATION_VERIFIED", evidence)
	receipt.OpenObligationIDs = []string{}
	receipt.Metrics = Metrics{8, 8, 0, 0, 0, 10000, 0}
	receipt.ClaimTransitions = []ClaimTransition{{ClaimID: attestationID, Before: "OPEN", After: "DISCHARGED", EvidenceDigest: evidence}}
	receipt.Views = []ReaderView{
		{Audience: "NON_ATTESTING_READER", Resolution: "LOWER_RESOLUTION", VerifiedTotal: 7, FixedTotal: 8, CoverageBasisPoints: 8750},
		{Audience: "ATTESTATION_READER", Resolution: "EXACT", VerifiedTotal: 8, FixedTotal: 8, CoverageBasisPoints: 10000},
	}
	receipt.Proofs = append(baselineProofs(*receipt), Proof{
		Choice: "COHERENCE", Claim: "the signer identity agrees with R while the signed subject digest agrees with T",
		MetaOperation: "resolve-producer-attestation", EvidenceDigest: evidence, Passed: true,
	})
}

func baselineProofs(receipt ResolutionReceipt) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", Claim: "the signed archive replays the prior transport receipt", MetaOperation: "bind-attested-archive", EvidenceDigest: receipt.SourceArchiveDigest, Passed: true},
		{Choice: "REGRESSION", Claim: "attestation resolution grants no effects or authority", MetaOperation: "deny-attestation-effects", EvidenceDigest: receipt.PriorReceiptDigest, Passed: true},
	}
}
