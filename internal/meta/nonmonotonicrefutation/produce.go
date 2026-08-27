package nonmonotonicrefutation

func Produce(sourcePath string, source []byte) ProducerReport {
	report := ProducerReport{
		Schema: ProducerSchema, Contract: CanonicalContract(), SourcePath: sourcePath,
		SourceDigest: DigestBytes(source), Producer: ProducerID, Consumer: ConsumerID,
		MetaOperation: MetaOperation, ProofChoice: ProofRegression,
		Effects: Effects{RepositoryWrites: 0, MutationAuthority: false},
		NotClaimed: []string{
			"truth of the domain claim outside the fixed fixtures",
			"probabilistic confidence or source credibility ranking",
			"automatic repository mutation or event-log replication",
		},
	}
	report.ReceiptDigest = DigestJSON(report)
	return report
}
