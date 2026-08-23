package languagereadiness

import (
	"fmt"

	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
)

func validateToolchainCLI(reports []metacli.Report, expectedHeadSHA string) (string, error) {
	if len(reports) != 1 {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain CLI evidence is not unique")
	}
	report := reports[0]
	if err := metacli.Validate(report, expectedHeadSHA); err != nil {
		return "", fmt.Errorf("verify toolchain CLI: %w", err)
	}
	if report.Decision != metacli.DecisionPass || report.Resolution != metacli.ResolutionExact {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain CLI decision %q/%q", report.Decision, report.Resolution)
	}
	return report.ReportDigest, nil
}
