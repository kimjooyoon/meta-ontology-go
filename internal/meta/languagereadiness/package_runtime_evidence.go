package languagereadiness

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagepackageruntime"
)

func validatePackageRuntime(
	reports []languagepackageruntime.Report, expectedHeadSHA string,
) (string, error) {
	if len(reports) == 0 {
		return "", nil
	}
	if len(reports) != 1 {
		return "", fmt.Errorf("FAIL_CLOSED: package runtime evidence is not unique")
	}
	report := reports[0]
	if err := languagepackageruntime.Validate(report, expectedHeadSHA); err != nil {
		return "", fmt.Errorf("verify language package runtime: %w", err)
	}
	if report.Decision != languagepackageruntime.DecisionPass ||
		report.Resolution != languagepackageruntime.ResolutionExact {
		return "", fmt.Errorf("FAIL_CLOSED: package runtime decision %q/%q",
			report.Decision, report.Resolution)
	}
	return report.ReportDigest, nil
}
