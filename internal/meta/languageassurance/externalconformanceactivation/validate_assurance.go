package externalconformanceactivation

func validAssurance(report assuranceReport) bool {
	if report.Schema != "gooo/language-assurance-report/v1" || report.SubjectSHA != PredecessorSHA ||
		report.DenominatorID != "gooo/language-assurance-denominator/v1" ||
		report.DenominatorDigest != "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02" ||
		report.AssuranceDecision != "PARTIAL" || report.CandidateDecision != "ALLOW_LIMITED" ||
		report.CandidateResolution != ResolutionExact || report.ReportDigest != AssuranceReportHash {
		return false
	}
	s := report.Summary
	if s.DenominatorTotal != 12 || s.Operating != 11 || s.NotImplemented != 1 ||
		s.ImplementationCoverageBPS != 9166 || s.UnknownTopDecisions != 0 ||
		s.UnresolvedIndicators != 0 || s.ViolatedGuardrails != 0 || s.RepositoryWrites != 0 ||
		len(report.Obligations) != 12 {
		return false
	}
	operating, target := 0, 0
	for _, item := range report.Obligations {
		if item.Status == "OPERATING" && item.Resolution == ResolutionExact && item.MetaOperation != "" {
			operating++
		}
		if item.MetricID == MetricID && item.Status == "NOT_IMPLEMENTED" && item.Resolution == "NONE" && item.MetaOperation == "" {
			target++
		}
	}
	return operating == 11 && target == 1
}
