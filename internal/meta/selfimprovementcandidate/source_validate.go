package selfimprovementcandidate

func sourceShapeKnown(source sourceObservation) bool {
	summary := source.Summary
	return source.Schema == "gooo/self-improvement-language-observation/v1" &&
		source.Metaprogram == "internal/meta/selfimprovementobservation" &&
		source.ContractID == "billing-operation-manifest-v2" &&
		source.Reason == "READ_ONLY_MINIMAL_VALUE_BOUND" &&
		source.Interpretation == "READ_ONLY_IMPROVEMENT_INPUT" &&
		coordinateEquals(summary.Coordinates, 16, 16) &&
		coordinateEquals(summary.SourceCoordinates, 15, 15) &&
		coordinateEquals(summary.Counterexamples, 6, 6) &&
		summary.GoooDefinitionFiles == 2 && summary.GoDefinitionFiles == 0 &&
		summary.ResourceSamples == 5 && summary.MaxWallMS > 0 &&
		summary.MaxRSSKiB > 0 && summary.BinaryBytes > 0 && summary.Unknowns == 0 &&
		validDigest(source.InputDigest) && sourceEvidenceKnown(source)
}

func sourceEvidenceKnown(source sourceObservation) bool {
	if len(source.Artifacts) != 3 || len(source.Indicators) != 16 ||
		len(source.Views) != 3 || len(source.Proofs) != 3 {
		return false
	}
	for _, artifact := range source.Artifacts {
		if artifact.Kind == "" || artifact.Schema == "" || artifact.Decision != "PASS" ||
			!validDigest(artifact.FileDigest) || !validEvidenceDigest(artifact.SemanticDigest) {
			return false
		}
	}
	for _, indicator := range source.Indicators {
		if indicator.ID == "" || indicator.Class == "" || indicator.ProofChoice == "" ||
			indicator.MetaOperation == "" || !indicator.Satisfied || indicator.Value != indicator.Target {
			return false
		}
	}
	return sourceViewsKnown(source.Views) && sourceProofsKnown(source.Proofs)
}

func sourceViewsKnown(views []sourceView) bool {
	expected := []struct {
		audience         string
		satisfied, total int
	}{{"USER", 5, 5}, {"TOOL_AUTHOR", 12, 12}, {"GOVERNOR", 16, 16}}
	for index, view := range views {
		item := expected[index]
		if view.Audience != item.audience || view.Satisfied != item.satisfied ||
			view.Total != item.total || view.BasisPoints != 10000 ||
			len(view.IndicatorIDs) != item.total {
			return false
		}
	}
	return true
}

func sourceProofsKnown(proofs []sourceProof) bool {
	for _, proof := range proofs {
		if proof.Choice == "" || proof.Claim == "" || proof.MetaOperation == "" ||
			!validDigest(proof.EvidenceDigest) || !proof.Passed {
			return false
		}
	}
	return true
}
