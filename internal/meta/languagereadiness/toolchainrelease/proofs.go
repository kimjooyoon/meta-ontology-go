package toolchainrelease

func buildProofs(corpusDigest, conceptDigest string, evidence []PlatformEvidence) []Proof {
	return []Proof{
		{
			ProofChoice: "FOUNDATION",
			Claim: "the v1 denominator fixes three stable x64 runner targets and twenty cases",
			EvidenceDigest: corpusDigest,
		},
		{
			ProofChoice: "COHERENCE",
			Claim: "native receipts and the release concept bind one exact source state",
			EvidenceDigest: conceptDigest,
		},
		{
			ProofChoice: "REGRESSION",
			Claim: "two binary and archive builds replay byte-equal for every target",
			EvidenceDigest: receiptSetDigest(evidence),
		},
	}
}
