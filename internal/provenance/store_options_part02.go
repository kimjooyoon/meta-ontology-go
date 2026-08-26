package provenance

func claimMatches(records []Evidence, semanticID, semanticDigest, graphDigest string) bool {
	for _, record := range records {
		if record.SemanticID == semanticID && record.Status == StatusVerified && record.SemanticDigest == semanticDigest && record.GraphDigest == graphDigest {
			return true
		}
	}
	return false
}
