package toolchainlsp

import "fmt"

func Validate(report Report, expectedHeadSHA string) error {
	if report.Schema != ReportSchema { return fmt.Errorf("toolchain lsp schema mismatch") }
	if report.HeadSHA != expectedHeadSHA || !validSHA(report.HeadSHA) { return fmt.Errorf("toolchain lsp head mismatch") }
	digest := report.ReportDigest; report.ReportDigest = ""
	if digestValue(report) != digest { return fmt.Errorf("toolchain lsp digest mismatch") }
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact || report.Reason != "TOOLCHAIN_LSP_READY" {
		return fmt.Errorf("toolchain lsp decision %s/%s", report.Decision, report.Resolution)
	}
	summary := report.Summary
	if summary.CasesSatisfied != 22 || summary.CasesTotal != 22 || summary.ReadinessBPS != 10000 ||
		summary.ProtocolCases != 16 || summary.CouplingCases != 6 || summary.AdvertisedCapabilities != 8 ||
		summary.ReadFeatures != 7 || summary.DiagnosticPaths != 3 || summary.NavigationPaths != 3 ||
		summary.SymbolPaths != 2 || summary.SemanticTokenPaths != 1 || summary.UTF16Replays != 1 ||
		summary.TranscriptReplays != 1 || summary.FailClosedPaths != 5 || summary.ConceptBindings != 1 ||
		summary.CodeBindings != 5 || summary.MetricBindings != 37 || summary.UseCaseBindings != 3 {
		return fmt.Errorf("toolchain lsp positive denominator mismatch")
	}
	if summary.MissingCases+summary.UnexpectedCases+summary.CaseFailures+summary.CapabilityGaps+
		summary.UnexpectedProtocolErrors+summary.DiagnosticGaps+summary.NonstandardWireFields+
		summary.StaleNavigationLeaks+summary.UnknownNavigationLeaks+summary.FailClosedNavigationLeaks+
		summary.Unresolved+summary.DigestFailures+summary.CorpusDrift+summary.ConceptDrift+
		summary.HeadMismatches+summary.ProofFailures+summary.RepositoryWrites+summary.MutationAuthorities != 0 {
		return fmt.Errorf("toolchain lsp guardrail is nonzero")
	}
	if len(report.Cases) != 22 || len(report.Indicators) != 37 || len(report.Proofs) != 3 { return fmt.Errorf("toolchain lsp evidence count mismatch") }
	for index, item := range report.Cases { if item.ID != caseContract[index].ID || item.Status != "SATISFIED" || item.EvidenceDigest == "" { return fmt.Errorf("toolchain lsp case %d failed", index) } }
	for _, item := range report.Indicators { if !item.Satisfied || item.Producer != "toolchainlsp.Evaluate" || item.Consumer != "self-improvement-cycle" || item.MetaOperation != MetaOperation { return fmt.Errorf("toolchain lsp indicator failed") } }
	for _, proof := range report.Proofs { if !proof.Passed { return fmt.Errorf("toolchain lsp proof failed") } }
	return nil
}
