package languagereadiness

import (
	"fmt"

	metaconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
)

func validateToolchainConformance(reports []metaconformance.Report,
	expectedHeadSHA string) (string, error) {
	if len(reports) != 1 {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain conformance evidence is not unique")
	}
	report := reports[0]
	if err := metaconformance.Validate(report, expectedHeadSHA); err != nil {
		return "", fmt.Errorf("verify toolchain conformance: %w", err)
	}
	if report.Decision != metaconformance.DecisionPass ||
		report.Resolution != metaconformance.ResolutionExact {
		return "", fmt.Errorf("FAIL_CLOSED: toolchain conformance decision %q/%q",
			report.Decision, report.Resolution)
	}
	return report.ReportDigest, nil
}
