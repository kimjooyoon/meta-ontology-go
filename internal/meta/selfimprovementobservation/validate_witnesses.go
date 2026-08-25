package selfimprovementobservation

func validDenominators(report SourceReport) bool {
	coordinates := report.Summary.Coordinates
	return coordinates.Satisfied == 15 && coordinates.Total == 15 && coordinates.BasisPoints == 10000 &&
		report.Summary.NotClaimed == 5 && report.Summary.Unknowns == 0 && validNotClaims(report.NotClaimed)
}

func validValue(value SourceValue) bool {
	return value.PrimaryArtifacts == 1 && value.ArtifactDigestChecks == 3 &&
		value.GoldenMatches == 1 && value.DeterministicReplays == 1
}

func validCompiler(compiler SourceCompiler) bool {
	return compiler.SourceFiles == 2 && compiler.GoooFiles == 2 && compiler.GoFiles == 0 &&
		compiler.GoooDefinitionBasisPoints == 10000 && compiler.RegisteredEmitters == 3
}

func validResources(resources SourceResources) bool {
	return resources.Samples == 5 && resources.ValidSamples == 5 && resources.MaxWallMS >= 0 &&
		resources.MaxRSSKiB > 0 && resources.BinaryBytes > 0 && resources.WallViolations == 0 &&
		resources.RSSViolations == 0 && resources.BinaryViolations == 0
}

func validCounterexamples(report CounterexampleReport) bool {
	return report.Schema == "gooo/language-example-counterexamples/v1" &&
		report.Satisfied == 6 && report.Total == 6
}

func validNotClaims(values []string) bool {
	expected := map[string]bool{
		"business correctness": true, "value-level computation": true,
		"production readiness":                                true,
		"performance beyond this runner and fixed sample set": true,
		"general-purpose code generation":                     true,
	}
	if len(values) != len(expected) {
		return false
	}
	for _, value := range values {
		if !expected[value] {
			return false
		}
		delete(expected, value)
	}
	return len(expected) == 0
}
