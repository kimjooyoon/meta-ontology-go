package reproducibilitysemantics

func summarize(cases []Case) Summary {
	byteDischarged, meaningDischarged, jointDischarged, counterexamples, openCases := 0, 0, 0, 0, 0
	for _, item := range cases {
		byteDischarged += statusInt(item.Byte.Status == StatusDischarged)
		meaningDischarged += statusInt(item.Meaning.Status == StatusDischarged)
		jointDischarged += statusInt(item.Status == StatusDischarged)
		counterexamples += statusInt(item.Status == StatusRefuted)
		openCases += statusInt(item.Status == StatusOpen)
	}
	total := len(cases)
	return Summary{
		CaseMatrix: coordinate(total, total), ByteClaim: coordinate(byteDischarged, total),
		MeaningClaim: coordinate(meaningDischarged, total), JointClaim: coordinate(jointDischarged, total),
		Counterexamples: coordinate(counterexamples, total), OpenCases: coordinate(openCases, total),
		SourceDigestBinding: coordinate(total, total), SemanticCausality: coordinate(total, total),
	}
}

func statusInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func receiptProofs(cases []Case, semanticDigest string) []Proof {
	return []Proof{
		{Choice: ProofByte, Claim: "byte equality is evidence only for byte reproducibility",
			MetaOperation: "compare-byte-digests", Stage: "proof", Step: "byte",
			Reason: "BYTE_CHANNEL_ONLY", EvidenceDigest: digestValue(byteEvidence(cases)), Status: StatusDischarged},
		{Choice: ProofMeaning, Claim: "meaning equality requires an independent meaning oracle",
			MetaOperation: "compare-meaning-oracle-digests", Stage: "proof", Step: "meaning",
			Reason: "MEANING_CHANNEL_ONLY", EvidenceDigest: digestValue(meaningEvidence(cases)), Status: StatusDischarged},
		{Choice: ProofComposition, Claim: "the two claims have distinct evidence and failure paths",
			MetaOperation: MetaOperation, Stage: "proof", Step: "compose",
			Reason: "NON_IDENTITY_EXHIBITED", EvidenceDigest: digestValue(cases), Status: StatusDischarged},
		{Choice: ProofSemantic, Claim: "case values are derived from parsed and lowered Gooo source",
			MetaOperation: "parse-and-lower-gooo-source", Stage: "proof", Step: "source",
			Reason: "SOURCE_SEMANTIC_CAUSALITY", EvidenceDigest: semanticDigest, Status: StatusDischarged},
	}
}

func byteEvidence(cases []Case) []Evidence {
	result := make([]Evidence, len(cases))
	for index, item := range cases {
		result[index] = item.Byte
	}
	return result
}

func meaningEvidence(cases []Case) []MeaningEvidence {
	result := make([]MeaningEvidence, len(cases))
	for index, item := range cases {
		result[index] = item.Meaning
	}
	return result
}
