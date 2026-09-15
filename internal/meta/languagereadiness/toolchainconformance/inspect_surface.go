package toolchainconformance

func inspectSurface(definition SurfaceDefinition, raw []byte, expectedHead string,
	summary *Summary) SurfaceResult {
	result := SurfaceResult{ID: definition.ID, Schema: definition.Schema,
		Status: "NOT_SATISFIED", EvidenceDigest: digestValue(raw)}
	before := blockingCount(*summary)
	envelope, err := decodeEnvelope(raw)
	if err != nil {
		summary.Unresolved++
		return result
	}
	result.HeadSHA = inspectIdentity(definition, envelope, expectedHead, summary)
	result.Cases = inspectCases(definition, envelope, summary)
	result.Indicators = inspectIndicators(definition, envelope, summary)
	result.Proofs = inspectProofs(definition, envelope, summary)
	if blockingCount(*summary) == before {
		result.Status = "SATISFIED"
		summary.SurfacesSatisfied++
	}
	return result
}
