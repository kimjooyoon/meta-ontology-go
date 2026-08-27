package experimentportfolio

func sealCausalityReport(report CausalityReport) CausalityReport {
	report.Digest = ""
	report.Digest = digestValue(report)
	return report
}

func causalityReportDigest(report CausalityReport) string {
	report.Digest = ""
	return digestValue(report)
}
