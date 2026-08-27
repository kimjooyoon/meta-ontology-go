package reproducibilitysemantics

func validateCaseProvenance(item Case) string {
	if item.Stage != "composition" || item.Step != "case" ||
		item.Byte.Producer != ProducerID || item.Byte.Consumer != ConsumerID ||
		item.Byte.MetaOperation != "compare-byte-digests" || item.Byte.ProofChoice != ProofByte ||
		item.Byte.Stage != "evidence" || item.Byte.Step != "byte" ||
		item.Meaning.Producer != ProducerID || item.Meaning.Consumer != ConsumerID ||
		item.Meaning.MetaOperation != "compare-meaning-oracle-digests" || item.Meaning.ProofChoice != ProofMeaning ||
		item.Meaning.Stage != "evidence" || item.Meaning.Step != "meaning" ||
		!validEvidenceDigest(item.Byte.Reference) || !validEvidenceDigest(item.Byte.Candidate) ||
		!validEvidenceDigest(item.Meaning.Expected) || !validEvidenceDigest(item.Meaning.Observed) {
		return "CASE_PROVENANCE_INVALID"
	}
	return ""
}

func judgeEvidence(reference, candidate string) (string, string) {
	if reference == "" || candidate == "" {
		return StatusOpen, "EVIDENCE_MISSING"
	}
	if reference == candidate {
		return StatusDischarged, "EVIDENCE_MATCH"
	}
	return StatusRefuted, "EVIDENCE_MISMATCH"
}
