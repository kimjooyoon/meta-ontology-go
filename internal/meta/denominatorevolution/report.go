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
	if report.Summary.CasesTotal == 0 {
		report.Summary = Summary{CasesTotal: CaseCount, FixedDenominatorDenominator: DenominatorSize, SourceCasesDenominator: CaseCount, PersistentClaimsDenominator: CaseCount, GuardrailObservationsDenominator: GuardrailCount, VersionRecordsDenominator: 2, V1NonretroactiveDenominator: 1}
	}
	report.Summary.Guardrails = makeGuardrails(forbiddenEstimateObserved(report.EmittedClaims), report.RepositorySnapshot.ChangedPaths, report.SourceProjection.ForbiddenPropositionPresent)
	report.Indicators = makeIndicators(report.Summary)
	if report.AggregateMetrics == nil {
		report.AggregateMetrics = []string{}
	}
	report.Digest = reportDigest(report)
	return report
}
