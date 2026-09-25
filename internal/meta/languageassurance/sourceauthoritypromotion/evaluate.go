package sourceauthoritypromotion

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func Evaluate(input Input) Report {
	report := Report{Schema: Schema, SubjectSHA: input.SubjectSHA,
		EligibilityDenominatorID: EligibilityDenominator, EligibilityDenominatorDigest: digest(Denominator()),
		Transition: Transition{MetricID: SourceMetric, MetaOperation: SourceOperation,
			FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE", EligibleStatus: "OPERATING", EligibleResolution: ResolutionExact}}
	assurance, upstream, decoded := decode(input)
	if !decoded {
		return finish(report, false, false, false, ReasonMalformed)
	}
	if assurance.SubjectSHA != input.SubjectSHA || upstream.SubjectSHA != input.SubjectSHA {
		return finish(report, false, false, false, ReasonSubjectMismatch)
	}
	baseline, baselineOK, baselineReason := validateAssurance(assurance)
	report.Baseline = baseline
	evidence, evidenceOK := validateUpstream(upstream)
	report.Evidence = evidence
	if !baselineOK {
		return finish(report, false, false, evidenceOK, baselineReason)
	}
	if !evidenceOK {
		return finish(report, false, true, false, ReasonUpstreamNotExact)
	}
	return finish(report, true, true, true, ReasonEligible)
}

func finish(report Report, eligible, baselineOK, evidenceOK bool, reason string) Report {
	report.Decision, report.Resolution, report.Enforcement = DecisionBlock, ResolutionInvariantOnly, DecisionBlock
	report.Reason = reason
	after, eligiblePaths, blockedPaths := report.Baseline.Operating, 0, 1
	if eligible {
		report.Decision, report.Resolution, report.Enforcement = DecisionEligible, ResolutionExact, EnforcementNoEffect
		after, eligiblePaths, blockedPaths = report.Baseline.Operating+1, 1, 0
	}
	report.Summary = Summary{DenominatorTotal: report.Baseline.Total, BeforeOperating: report.Baseline.Operating,
		AfterOperating: after, BeforeCoverageBPS: coverage(report.Baseline.Operating, report.Baseline.Total),
		AfterCoverageBPS: coverage(after, report.Baseline.Total), EligiblePaths: eligiblePaths, BlockedPaths: blockedPaths}
	report.Indicators = buildIndicators(baselineOK, evidenceOK, eligible)
	report.ReportDigest = digest(report)
	return report
}

func coverage(value, total int) int {
	if total == 0 {
		return 0
	}
	return value * 10000 / total
}

func digest(value any) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}
