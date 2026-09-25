package languagediagnosticprovenance

import "fmt"

func Validate(report Report, expectedHeadSHA string) error {
	expected := report
	expected.ReportDigest = ""
	switch {
	case report.Schema != ReportSchema:
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance schema mismatch")
	case report.Source.ExpectedHeadSHA != expectedHeadSHA || expectedHeadSHA == "":
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance head mismatch")
	case report.Decision != DecisionPass || report.Resolution != ResolutionExact:
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance decision %s/%s",
			report.Decision, report.Resolution)
	case report.Source.MetaOperation != ExpectedMetaOperation:
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance meta operation mismatch")
	case report.Source.RegistryDigest != digestJSON(Registry()):
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance registry mismatch")
	case !exactSummary(report.Summary):
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance summary mismatch")
	case !allStagesPassed(report.Stages) || !allProofsPassed(report.Proofs):
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance proof mismatch")
	case report.RepositoryWrites != 0 || report.MutationAuthorized:
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance effect boundary mismatch")
	case report.ReportDigest != digestJSON(expected):
		return fmt.Errorf("FAIL_CLOSED: diagnostic provenance digest mismatch")
	}
	return nil
}
