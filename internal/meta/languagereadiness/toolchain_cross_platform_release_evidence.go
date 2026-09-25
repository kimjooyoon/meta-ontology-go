package languagereadiness

import (
	"fmt"

	release "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
)

func validateToolchainCrossPlatformRelease(reports []release.Report, expectedHeadSHA string) (string, error) {
	if len(reports) != 1 {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain cross-platform release evidence is not unique")
	}
	report := reports[0]
	if err := release.Validate(report, expectedHeadSHA); err != nil {
		return "", fmt.Errorf("verify toolchain cross-platform release: %w", err)
	}
	if report.Decision != release.DecisionPass || report.Resolution != release.ResolutionExact {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain release decision %q/%q", report.Decision, report.Resolution)
	}
	return report.ReportDigest, nil
}
