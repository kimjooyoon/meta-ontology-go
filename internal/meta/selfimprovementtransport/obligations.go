package selfimprovementtransport

type evaluationInput struct {
	contract            ContractEvidence
	contractErr         error
	expectedRepository  string
	expectedRunID       int64
	source              observationHeader
	sourceErr           error
	observationRaw      []byte
	producer            ProducerReceipt
	producerErr         error
	metadata            TransportMetadata
	metadataErr         error
	actualArchiveDigest string
}

func evaluateObligations(input evaluationInput) []Obligation {
	return []Obligation{
		sourceRepositoryObligation(input),
		checkoutBindingObligation(input),
		producerIdentityObligation(input),
		logicalDigestObligation(input),
		immutableLocatorObligation(input),
		archiveDigestObligation(input),
		consumerReplayObligation(input),
		attestationObligationFor(input.metadata.Attestation),
	}
}
