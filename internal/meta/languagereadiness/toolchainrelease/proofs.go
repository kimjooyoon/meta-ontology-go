package toolchainrelease

import "fmt"

func buildProofs(corpus Corpus, corpusDigest, conceptDigest string, evidence []PlatformEvidence) []Proof {
	return []Proof{
		{
			ProofChoice:    "FOUNDATION",
			Claim:          fmt.Sprintf("the v%d corpus declares %d native runner targets and %d release cases", corpus.Version, len(corpus.Targets), len(corpus.Cases)),
			EvidenceDigest: corpusDigest,
		},
		{
			ProofChoice:    "COHERENCE",
			Claim:          "native receipts and the release concept bind one exact source state",
			EvidenceDigest: conceptDigest,
		},
		{
			ProofChoice:    "REGRESSION",
			Claim:          "two binary and archive builds replay byte-equal for every target",
			EvidenceDigest: receiptSetDigest(evidence),
		},
	}
}
