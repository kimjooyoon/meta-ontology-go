package selfimprovementcandidate

func validSource(head string, runID int64) sourceObservation {
	digest := fixtureDigest()
	source := sourceObservation{Schema: "gooo/self-improvement-language-observation/v1",
		Metaprogram: "internal/meta/selfimprovementobservation", SubjectSHA: head,
		SourceWorkflowRunID: runID, ContractID: "billing-operation-manifest-v2",
		Decision: "OBSERVED", Resolution: ResolutionExact, Reason: "READ_ONLY_MINIMAL_VALUE_BOUND",
		Interpretation: "READ_ONLY_IMPROVEMENT_INPUT", InputDigest: digest,
		Summary: sourceSummary{Coordinates: coordinate(16, 16), SourceCoordinates: coordinate(15, 15),
			Counterexamples: coordinate(6, 6), GoooDefinitionFiles: 2, GoDefinitionFiles: 0,
			ResourceSamples: 5, MaxWallMS: 12, MaxRSSKiB: 10712, BinaryBytes: 12529067},
		NotClaimed: append([]string{}, expectedSourceNonClaims...)}
	for range 3 {
		source.Artifacts = append(source.Artifacts, sourceArtifact{Kind: "artifact", Schema: "schema",
			FileDigest: digest, SemanticDigest: digest, Decision: "PASS"})
	}
	for range 16 {
		source.Indicators = append(source.Indicators, sourceIndicator{ID: "indicator", Class: "DRIVER",
			ProofChoice: "FOUNDATION", MetaOperation: "observe", Value: 1, Target: 1, Satisfied: true})
	}
	source.Views = []sourceView{
		fixtureView("USER", 5), fixtureView("TOOL_AUTHOR", 12), fixtureView("GOVERNOR", 16),
	}
	for _, choice := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		source.Proofs = append(source.Proofs, sourceProof{Choice: choice, Claim: "claim",
			MetaOperation: "observe", EvidenceDigest: digest, Passed: true})
	}
	source.Digest = sourceDigest(source)
	return source
}

func fixtureView(audience string, total int) sourceView {
	identifiers := make([]string, total)
	for index := range identifiers {
		identifiers[index] = "indicator"
	}
	return sourceView{Audience: audience, Resolution: "view", Satisfied: total,
		Total: total, BasisPoints: 10000, IndicatorIDs: identifiers}
}

func reseal(source *sourceObservation) { source.Digest = sourceDigest(*source) }
