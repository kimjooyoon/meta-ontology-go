package languagereadiness

import (
	"fmt"

	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
)

func validateToolchainFormatFix(reports []metaff.Report, expectedHeadSHA string) (string, error) {
	if len(reports) != 1 {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain format/fix evidence is not unique")
	}
	report := reports[0]
	if err := metaff.Validate(report, expectedHeadSHA); err != nil {
		return "", fmt.Errorf("verify toolchain format/fix: %w", err)
	}
	if report.Decision != metaff.DecisionPass || report.Resolution != metaff.ResolutionExact {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain format/fix decision %q/%q",
			report.Decision, report.Resolution)
	}
	return report.ReportDigest, nil
}
