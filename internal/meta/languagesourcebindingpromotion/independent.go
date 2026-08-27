package languagesourcebindingpromotion

func assessIndependent(oracleRaw, receiptRaw []byte, head string) component {
	if len(oracleRaw) == 0 {
		return open("ARTIFACT_ORACLE_EVIDENCE_MISSING", "PROMOTION_INPUT", "read-oracle")
	}
	oracle, err := decodeStrict[oracleEnvelope](oracleRaw)
	if err != nil || !verifyOracleDigest(oracle) {
		return refuted("ARTIFACT_ORACLE_EVIDENCE_INVALID", "PROMOTION_INPUT", "read-oracle")
	}
	if oracle.Decision != DecisionPass {
		if oracle.Decision == "UNKNOWN" {
			return open("ARTIFACT_ORACLE_DECISION_UNKNOWN", "PROMOTION_COMPARE", "oracle-decision", oracle.Digest)
		}
		return refuted("ARTIFACT_ORACLE_DECISION_REJECTED", "PROMOTION_COMPARE", "oracle-decision", oracle.Digest)
	}
	if oracle.Schema != "gooo/language-artifact-oracle/v1" || oracle.Scope != "SOURCE_EXECUTION_ARTIFACT_BINDING_ONLY" ||
		oracle.HeadSHA != head || oracle.Resolution != ResolutionExact || oracle.RepositoryWrites != 0 || oracle.MutationAuthority {
		return refuted("ARTIFACT_ORACLE_EVIDENCE_INVALID", "PROMOTION_COMPARE", "oracle-identity", oracle.Digest)
	}
	summary, err := decodeView[oracleSummary](oracle.Summary)
	if err != nil || summary.CasesSatisfied != 4 || summary.CasesTotal != 4 ||
		summary.ProducerDependencies != 0 || summary.SemanticCorrectnessClaims != 0 {
		return refuted("ARTIFACT_ORACLE_SUMMARY_MISMATCH", "PROMOTION_COMPARE", "oracle-summary", oracle.Digest)
	}
	receipt, err := decodeStrict[receiptEnvelope](receiptRaw)
	if err != nil || !verifyReceiptDigest(receipt) {
		return refuted("SOURCE_EXECUTION_RECEIPT_INVALID", "PROMOTION_INPUT", "read-receipt", oracle.Digest)
	}
	cases, err := decodeView[[]oracleCase](oracle.Cases)
	if err != nil {
		return refuted("ARTIFACT_ORACLE_CASES_INVALID", "PROMOTION_COMPARE", "oracle-cases", oracle.Digest)
	}
	for _, item := range cases {
		if item.ID != "genuine-source-bound" {
			continue
		}
		if item.Status != "SATISFIED" || item.ObservedDecision != DecisionPass || item.ObservedResolution != ResolutionExact {
			return refuted("ARTIFACT_ORACLE_GENUINE_CASE_REJECTED", "PROMOTION_COMPARE", "genuine-source-bound", oracle.Digest)
		}
		if item.ArtifactDigest != receipt.Digest || item.SourceDigest != receipt.SourceDigest {
			return refuted("SOURCE_BINDING_EVIDENCE_LINK_MISMATCH", "PROMOTION_LINK", "receipt-digest", oracle.Digest, receipt.Digest)
		}
		return discharged("INDEPENDENT_SOURCE_BINDING_EXACT", "PROMOTION_INPUT", "artifact-oracle", oracle.Digest, receipt.Digest)
	}
	return refuted("ARTIFACT_ORACLE_GENUINE_CASE_MISSING", "PROMOTION_COMPARE", "genuine-source-bound", oracle.Digest)
}
