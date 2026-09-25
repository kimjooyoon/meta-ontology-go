package languagedelivery

func inspectLanguageTest(data []byte, head string, receipt *LanguageTestReceipt, entry ManifestEntry) SourceObservation {
	if err := unmarshalReceipt(data, receipt); err != nil {
		return unknownObservation(SourceTest, entry, "SOURCE_JSON_UNKNOWN")
	}
	observation := baseObservation(SourceTest, entry, receipt.Schema, receipt.Decision, receipt.Resolution)
	observation.RepositoryWrites = receipt.RepositoryWrites + receipt.Summary.Effects.RepositoryWrites
	observation.MutationAuthority = receipt.MutationAuthority || receipt.Summary.Effects.MutationAuthority
	if receipt.Schema != languageTestReportSchema {
		return finalizeObservation(observation, receipt.Schema, languageTestReportSchema)
	}
	if receipt.SubjectSHA != head {
		return headUnknown(observation)
	}
	if receipt.Summary.Unknowns != 0 || receipt.Decision == "UNKNOWN" || receipt.Resolution == "LOWER_RESOLUTION" {
		observation.State, observation.Reason = "UNKNOWN", "LANGUAGE_TEST_RECEIPT_UNKNOWN"
		return observation
	}
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" {
		return finalizeObservation(observation, receipt.Schema, languageTestReportSchema)
	}
	if !languageTestBoundaryExact(*receipt) {
		observation.State, observation.Reason = "FAIL", "LANGUAGE_TEST_BOUNDARY_NOT_EXACT"
		return observation
	}
	return finalizeObservation(observation, receipt.Schema, languageTestReportSchema)
}

func languageTestBoundaryExact(receipt LanguageTestReceipt) bool {
	coordinates := receipt.Summary.Coordinates
	compiler := receipt.Summary.Compiler
	return coordinates.Satisfied == 12 && coordinates.Total == 12 && coordinates.BasisPoints == 10000 &&
		receipt.Summary.DeclaredTests == 2 && receipt.Summary.ExecutedTests == 2 && receipt.Summary.PassedTests == 2 &&
		receipt.Summary.ReceiptDigestVariants == 1 && receipt.Summary.ExecutionDigestVariants == 1 &&
		receipt.Summary.AssertionRejections == 1 && receipt.Summary.MissingTestRejections == 1 &&
		compiler.Go127Runtimes == 2 && len(compiler.ExecutableDigest) == 71 && compiler.ExecutableDigest[:7] == "sha256:" &&
		receipt.Summary.NonClaims == 3 && languageTestViewsExact(receipt.Views)
}

func languageTestViewsExact(views []LanguageTestView) bool {
	want := []LanguageTestView{
		{Audience: "USER", Resolution: "USER_VISIBLE", Satisfied: 4, Total: 4, BasisPoints: 10000},
		{Audience: "TOOL_AUTHOR", Resolution: "TOOL_CONTRACT", Satisfied: 8, Total: 8, BasisPoints: 10000},
		{Audience: "GOVERNOR", Resolution: "FULL_RECEIPT", Satisfied: 12, Total: 12, BasisPoints: 10000},
	}
	if len(views) != len(want) {
		return false
	}
	for index := range want {
		if views[index] != want[index] {
			return false
		}
	}
	return true
}
