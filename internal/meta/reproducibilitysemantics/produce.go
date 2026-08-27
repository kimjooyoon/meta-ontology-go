package reproducibilitysemantics

func Produce(sourcePath, headSHA string, source []byte) Receipt {
	declared, semanticDigest, err := deriveProducerCases(sourcePath, source)
	if err != nil {
		return Receipt{Schema: ReceiptSchema, Version: 1, ContractID: ContractID,
			SourcePath: sourcePath, SourceDigest: digestBytes(source), HeadSHA: headSHA,
			Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperation,
			ProofChoice: ProofComposition, Stage: "receipt", Step: "produce",
			Reason: "SOURCE_DERIVATION_FAILED", Authority: Authority{}}
	}
	cases := make([]Case, len(declared))
	for index, item := range declared {
		cases[index] = caseFromDeclared(item)
	}
	for index := range cases {
		cases[index].Byte.Producer = ProducerID
		cases[index].Byte.Consumer = ConsumerID
		cases[index].Meaning.Producer = ProducerID
		cases[index].Meaning.Consumer = ConsumerID
		cases[index].Status, cases[index].Reason = compose(cases[index].Byte.Status, cases[index].Meaning.Status)
	}
	return sealReceipt(Receipt{
		Schema: ReceiptSchema, Version: 1, ContractID: ContractID,
		SourcePath: sourcePath, SourceDigest: digestBytes(source), SemanticDigest: semanticDigest, HeadSHA: headSHA,
		Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperation,
		ProofChoice: ProofComposition, Stage: "receipt", Step: "produce",
		Reason: "CLAIM_CHANNELS_SEPARATED", Cases: cases,
		Summary: summarize(cases), Proofs: receiptProofs(cases, semanticDigest), Authority: Authority{},
	})
}

func caseFromDeclared(item declaredCase) Case {
	byteStatus, byteReason := compare(item.ByteReference, item.ByteCandidate)
	if item.ByteReference == "" || item.ByteCandidate == "" {
		byteStatus, byteReason = StatusOpen, "BYTE_EVIDENCE_MISSING"
	}
	meaningStatus, meaningReason := compare(item.MeaningExpected, item.MeaningObserved)
	if item.MeaningExpected == "" || item.MeaningObserved == "" {
		meaningStatus, meaningReason = StatusOpen, "MEANING_EVIDENCE_MISSING"
	}
	return Case{
		ID: item.ID,
		Byte: Evidence{MetaOperation: "compare-byte-digests", ProofChoice: ProofByte,
			Stage: "evidence", Step: "byte", Reason: byteReason,
			Reference: digestText(item.ByteReference), Candidate: digestText(item.ByteCandidate), Status: byteStatus},
		Meaning: MeaningEvidence{MetaOperation: "compare-meaning-oracle-digests", ProofChoice: ProofMeaning,
			Stage: "evidence", Step: "meaning", Reason: meaningReason,
			Expected: digestText(item.MeaningExpected), Observed: digestText(item.MeaningObserved), Status: meaningStatus},
		Stage: "composition", Step: "case", Reason: "CLAIM_CHANNELS_PENDING",
	}
}
