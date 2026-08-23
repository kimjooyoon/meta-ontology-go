package languagereadiness

import (
	"fmt"

	metalsp "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainlsp"
)

func validateToolchainLSP(reports []metalsp.Report, expectedHeadSHA string) (string, error) {
	if len(reports) != 1 { return "", fmt.Errorf("FAIL_CLOSED: toolchain LSP evidence is not unique") }
	report := reports[0]
	if err := metalsp.Validate(report, expectedHeadSHA); err != nil { return "", fmt.Errorf("verify toolchain LSP: %w", err) }
	if report.Decision != metalsp.DecisionPass || report.Resolution != metalsp.ResolutionExact {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain LSP decision %q/%q", report.Decision, report.Resolution)
	}
	return report.ReportDigest, nil
}
