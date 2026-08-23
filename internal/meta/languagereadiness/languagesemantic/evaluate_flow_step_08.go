package languagesemantic

func evaluateFlowStep07(flow *evaluateFlowState) {
	if err := validateSyntaxReceipt(flow.slot05, flow.slot00.ExpectedHeadSHA); err != nil {
		report := unresolvedReport(flow.slot01, flow.slot00.ExpectedHeadSHA, digestBytes(flow.slot02), err.Error())
		report.Source.SyntaxArtifactDigest = digestBytes(flow.slot04)
		report.Source.SyntaxReportDigest = flow.slot05.ReportDigest
		finalizeReport(&report)
		{
			flow.result0, flow.result1 = report, nil
			flow.done = true
			return
		}
	}
}
