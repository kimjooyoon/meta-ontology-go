package denominatorevolution

import (
	"encoding/json"
	"os"
)

func WriteReport(filename string, report Report) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(filename, raw, 0o644)
}

func finishFailure(report Report, reason, resolution string) Report {
	report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", resolution, reason
	report.Summary = Summary{CasesTotal: CaseCount, FixedDenominatorDenominator: DenominatorSize, ForbiddenEstimateDenominator: 1}
	report.Indicators = []Indicator{}
	report.Digest = reportDigest(report)
	return report
}
