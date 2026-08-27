package reproducibilitysemantics

func Produce(sourcePath, headSHA string, source []byte) Receipt {
	cases := []Case{
		makeCase("both-discharged", "artifact/canonical/approved/v1", "artifact/canonical/approved/v1", "meaning/charge-and-ledger/v1"),
		makeCase("reproducible-but-wrong", "artifact/canonical/approved/v1", "artifact/canonical/approved/v1", "meaning/render-approved/v1"),
		makeCase("meaningful-but-unreproduced", "artifact/canonical/approved/v1", "artifact/canonical/approved/v2", "meaning/charge-and-ledger/v1"),
		makeCase("claims-open", "", "", ""),
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
		SourcePath: sourcePath, SourceDigest: digestBytes(source), HeadSHA: headSHA,
		Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperation,
		ProofChoice: ProofComposition, Stage: "receipt", Step: "produce",
		Reason: "CLAIM_CHANNELS_SEPARATED", Cases: cases,
		Summary: summarize(cases), Proofs: receiptProofs(cases), Authority: Authority{},
	})
}

func makeCase(id, reference, candidate, meaning string) Case {
	byteStatus, byteReason := compare(reference, candidate)
	if reference == "" || candidate == "" {
		byteStatus, byteReason = StatusOpen, "BYTE_EVIDENCE_MISSING"
	}
	expectedMeaning := "meaning/charge-and-ledger/v1"
	meaningStatus, meaningReason := compare(expectedMeaning, meaning)
	if meaning == "" {
		meaningStatus, meaningReason = StatusOpen, "MEANING_EVIDENCE_MISSING"
		expectedMeaning = ""
	}
	return Case{
		ID: id,
		Byte: Evidence{MetaOperation: "compare-byte-digests", ProofChoice: ProofByte,
			Stage: "evidence", Step: "byte", Reason: byteReason,
			Reference: digestText(reference), Candidate: digestText(candidate), Status: byteStatus},
		Meaning: MeaningEvidence{MetaOperation: "compare-meaning-oracle-digests", ProofChoice: ProofMeaning,
			Stage: "evidence", Step: "meaning", Reason: meaningReason,
			Expected: digestText(expectedMeaning), Observed: digestText(meaning), Status: meaningStatus},
		Stage: "composition", Step: "case", Reason: "CLAIM_CHANNELS_PENDING",
	}
}

func digestText(value string) string {
	if value == "" {
		return ""
	}
	return digestBytes([]byte(value))
}

func compare(reference, candidate string) (string, string) {
	if reference == candidate {
		return StatusDischarged, "EVIDENCE_MATCH"
	}
	return StatusRefuted, "EVIDENCE_MISMATCH"
}

func compose(byteStatus, meaningStatus string) (string, string) {
	if byteStatus == StatusDischarged && meaningStatus == StatusDischarged {
		return StatusDischarged, "BOTH_CLAIMS_DISCHARGED"
	}
	if byteStatus == StatusDischarged && meaningStatus == StatusRefuted {
		return StatusRefuted, "REPRODUCIBLE_BUT_WRONG"
	}
	if byteStatus == StatusRefuted && meaningStatus == StatusDischarged {
		return StatusRefuted, "MEANINGFUL_BUT_UNREPRODUCED"
	}
	return StatusOpen, "CLAIMS_OPEN"
}
