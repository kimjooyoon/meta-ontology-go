package reproducibilitysemantics

func validateReceiptProofs(receipt Receipt) string {
	if receipt.Proofs[0].Choice != ProofByte || receipt.Proofs[0].Claim != "byte equality is evidence only for byte reproducibility" || receipt.Proofs[0].MetaOperation != "compare-byte-digests" ||
		receipt.Proofs[0].Stage != "proof" || receipt.Proofs[0].Step != "byte" || receipt.Proofs[0].Reason != "BYTE_CHANNEL_ONLY" || receipt.Proofs[0].Status != StatusDischarged ||
		receipt.Proofs[0].EvidenceDigest != digestValue(judgeByteEvidence(receipt.Cases)) {
		return "BYTE_PROOF_INVALID"
	}
	if receipt.Proofs[1].Choice != ProofMeaning || receipt.Proofs[1].Claim != "meaning equality requires an independent meaning oracle" || receipt.Proofs[1].MetaOperation != "compare-meaning-oracle-digests" ||
		receipt.Proofs[1].Stage != "proof" || receipt.Proofs[1].Step != "meaning" || receipt.Proofs[1].Reason != "MEANING_CHANNEL_ONLY" || receipt.Proofs[1].Status != StatusDischarged ||
		receipt.Proofs[1].EvidenceDigest != digestValue(judgeMeaningEvidence(receipt.Cases)) {
		return "MEANING_PROOF_INVALID"
	}
	if receipt.Proofs[2].Choice != ProofComposition || receipt.Proofs[2].Claim != "the two claims have distinct evidence and failure paths" || receipt.Proofs[2].MetaOperation != MetaOperation ||
		receipt.Proofs[2].Stage != "proof" || receipt.Proofs[2].Step != "compose" || receipt.Proofs[2].Reason != "NON_IDENTITY_EXHIBITED" || receipt.Proofs[2].Status != StatusDischarged ||
		receipt.Proofs[2].EvidenceDigest != digestValue(receipt.Cases) {
		return "COMPOSITION_PROOF_INVALID"
	}
	if receipt.Proofs[3].Choice != ProofSemantic || receipt.Proofs[3].Claim != "case values are derived from parsed and lowered Gooo source" ||
		receipt.Proofs[3].MetaOperation != "parse-and-lower-gooo-source" || receipt.Proofs[3].Stage != "proof" ||
		receipt.Proofs[3].Step != "source" || receipt.Proofs[3].Reason != "SOURCE_SEMANTIC_CAUSALITY" ||
		receipt.Proofs[3].Status != StatusDischarged || receipt.Proofs[3].EvidenceDigest != receipt.SemanticDigest {
		return "SEMANTIC_PROOF_INVALID"
	}
	return ""
}

func judgeByteEvidence(cases []Case) []Evidence {
	result := make([]Evidence, len(cases))
	for index, item := range cases {
		result[index] = item.Byte
	}
	return result
}

func judgeMeaningEvidence(cases []Case) []MeaningEvidence {
	result := make([]MeaningEvidence, len(cases))
	for index, item := range cases {
		result[index] = item.Meaning
	}
	return result
}

func judgeProofs(receipt Receipt, judgment Judgment) []Proof {
	return []Proof{
		{Choice: ProofByte, Claim: "consumer recomputed byte equality independently", MetaOperation: "compare-byte-digests",
			Stage: "judge", Step: "byte", Reason: "BYTE_REPLAY_INDEPENDENT", EvidenceDigest: digestValue(judgment.Cases), Status: StatusDischarged},
		{Choice: ProofMeaning, Claim: "consumer recomputed meaning equality independently", MetaOperation: "compare-meaning-oracle-digests",
			Stage: "judge", Step: "meaning", Reason: "MEANING_REPLAY_INDEPENDENT", EvidenceDigest: digestValue(judgment.Cases), Status: StatusDischarged},
		{Choice: ProofComposition, Claim: "consumer preserved the two failure paths and four-case matrix", MetaOperation: MetaOperation,
			Stage: "judge", Step: "compose", Reason: "MATRIX_REPLAY_INDEPENDENT", EvidenceDigest: receipt.ReceiptDigest, Status: StatusDischarged},
		{Choice: ProofSemantic, Claim: "consumer replayed parsed and lowered Gooo source", MetaOperation: "parse-and-lower-gooo-source",
			Stage: "judge", Step: "source", Reason: "SOURCE_REPLAY_INDEPENDENT", EvidenceDigest: judgment.SemanticDigest, Status: StatusDischarged},
	}
}
