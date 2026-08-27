package externaloraclehumility

func ProduceSourceReceipt(subject, sourcePath string, source []byte, contract Contract) SourceReceipt {
	claims := append([]Claim(nil), contract.Source.Claims...)
	return SourceReceipt{
		Schema: ReceiptSchema, SubjectSHA: subject, SourcePath: sourcePath,
		SourceSHA256: DigestBytes(source), Producer: "source-receipt-producer",
		Consumer: "independent-judge", MetaOperation: "emit-source-receipt",
		ProofChoice: "FOUNDATION", Stage: "observe", Step: "source-receipt",
		Reason: "GOOO_SOURCE_INTENT_BOUND", Claims: claims,
	}
}
