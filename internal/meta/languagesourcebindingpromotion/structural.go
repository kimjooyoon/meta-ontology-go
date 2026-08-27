package languagesourcebindingpromotion

func assessStructural(producerRaw, receiptRaw []byte, head string) component {
	producer, err := decodeStrict[producerEnvelope](producerRaw)
	if err != nil || !verifyProducerDigest(producer) {
		return refuted("SOURCE_EXECUTION_EVIDENCE_INVALID", "PROMOTION_INPUT", "read-producer")
	}
	if producer.Decision != DecisionPass {
		if producer.Decision == "UNKNOWN" {
			return open("SOURCE_EXECUTION_DECISION_UNKNOWN", "PROMOTION_COMPARE", "producer-decision", producer.Digest)
		}
		return refuted("SOURCE_EXECUTION_DECISION_REJECTED", "PROMOTION_COMPARE", "producer-decision", producer.Digest)
	}
	if producer.Schema != "gooo/language-source-execution-artifact/v1" || producer.HeadSHA != head ||
		producer.Resolution != ResolutionExact || producer.RepositoryWrites != 0 || producer.MutationAuthority {
		return refuted("SOURCE_EXECUTION_EVIDENCE_INVALID", "PROMOTION_COMPARE", "producer-identity", producer.Digest)
	}
	receipt, err := decodeStrict[receiptEnvelope](receiptRaw)
	if err != nil || !verifyReceiptDigest(receipt) || receipt.Schema != "gooo/source-execution-receipt/v1" ||
		receipt.Decision != DecisionPass || receipt.Resolution != ResolutionExact {
		return refuted("SOURCE_EXECUTION_RECEIPT_INVALID", "PROMOTION_INPUT", "read-receipt", producer.Digest)
	}
	summary, err := decodeView[producerSummary](producer.Summary)
	if err != nil || summary.CasesSatisfied != 4 || summary.CasesTotal != 4 {
		return refuted("SOURCE_EXECUTION_SUMMARY_MISMATCH", "PROMOTION_COMPARE", "producer-summary", producer.Digest)
	}
	cases, err := decodeView[[]producerCase](producer.Cases)
	if err != nil {
		return refuted("SOURCE_EXECUTION_CASES_INVALID", "PROMOTION_COMPARE", "producer-cases", producer.Digest)
	}
	for _, item := range cases {
		if item.ID == "execute-billing" && item.Status == "SATISFIED" && item.EvidenceDigest == receipt.Digest {
			return discharged("SOURCE_EXECUTION_EVIDENCE_EXACT", "PROMOTION_INPUT", "source-execution", producer.Digest, receipt.Digest)
		}
	}
	return refuted("SOURCE_EXECUTION_RECEIPT_LINK_MISMATCH", "PROMOTION_LINK", "producer-receipt", producer.Digest, receipt.Digest)
}
