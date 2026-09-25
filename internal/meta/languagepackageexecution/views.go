package languagepackageexecution

func views(report Report) []AudienceView {
	user := []string{"PACKAGE_FIXED_CASES", "PACKAGE_EXECUTIONS", "PACKAGE_DIAGNOSTIC_REJECTIONS"}
	tool := append(append([]string{}, user...), "PACKAGE_SOURCE_FILES", "PACKAGE_DETERMINISTIC_REPLAYS", "PACKAGE_EVENTS")
	governor := make([]string, 0, len(report.Indicators))
	for _, item := range report.Indicators {
		governor = append(governor, item.ID)
	}
	return []AudienceView{
		{Audience: "USER", ReaderResolution: "OUTCOME", VisibleFacts: user, HiddenFacts: len(governor) - len(user), FactsDigest: report.FactsDigest},
		{Audience: "TOOL_AUTHOR", ReaderResolution: "OPERATION", VisibleFacts: tool, HiddenFacts: len(governor) - len(tool), FactsDigest: report.FactsDigest},
		{Audience: "GOVERNOR", ReaderResolution: "PROOF", VisibleFacts: governor, HiddenFacts: 0, FactsDigest: report.FactsDigest},
	}
}
