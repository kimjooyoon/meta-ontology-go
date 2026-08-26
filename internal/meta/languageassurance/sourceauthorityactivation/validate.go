package sourceauthorityactivation

import (
	"encoding/hex"
	"encoding/json"

	promotion "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthoritypromotion"
)

func validateEvidence(input Input) (promotion.Report, int, string, string) {
	bindings := artifactBindings(input)
	exact := 0
	for _, binding := range bindings {
		if binding.Exact {
			exact++
		}
	}
	if !validSHA(input.SubjectSHA) || len(input.Assurance) == 0 || len(input.Upstream) == 0 || len(input.Eligibility) == 0 {
		return promotion.Report{}, exact, ReasonUnavailable, ResolutionUnknown
	}
	if exact != len(bindings) {
		return promotion.Report{}, exact, ReasonDigestMismatch, ResolutionInvariant
	}
	var report promotion.Report
	if err := json.Unmarshal(input.Eligibility, &report); err != nil {
		return promotion.Report{}, exact, ReasonUnavailable, ResolutionUnknown
	}
	if !validEligibility(report) {
		return report, exact, ReasonEligibility, ResolutionInvariant
	}
	return report, exact, "", ResolutionExact
}

func validEligibility(report promotion.Report) bool {
	copy := report
	copy.ReportDigest = ""
	if report.Schema != promotion.Schema || report.SubjectSHA != PredecessorSHA || report.Decision != promotion.DecisionEligible || report.Resolution != promotion.ResolutionExact || report.Enforcement != promotion.EnforcementNoEffect || report.Reason != promotion.ReasonEligible {
		return false
	}
	if report.ReportDigest != EligibilityReportHash || digestValue(copy) != EligibilityReportHash || report.RepositoryWrites != 0 || report.PromotionApplied != 0 {
		return false
	}
	if report.Transition.MetricID != MetricID || report.Transition.MetaOperation != MetaOperation || report.Transition.FromStatus != "NOT_IMPLEMENTED" || report.Transition.FromResolution != "NONE" || report.Transition.EligibleStatus != "OPERATING" || report.Transition.EligibleResolution != ResolutionExact {
		return false
	}
	if report.Baseline.Total != 12 || report.Baseline.Operating != 6 || report.Baseline.NotImplemented != 6 || report.Baseline.CoverageBPS != 5000 || report.Summary.AfterOperating != 7 || report.Summary.AfterCoverageBPS != 5833 {
		return false
	}
	return report.Evidence.CasesPassed == 3 && report.Evidence.CasesTotal == 3 && report.Evidence.CoverageBPS == 10000 && validIndicatorSplit(report.Indicators)
}

func validIndicatorSplit(indicators []promotion.Indicator) bool {
	counts := map[string]int{}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return false
		}
		counts[indicator.Class]++
		counts[indicator.ProofChoice]++
	}
	return len(indicators) == 6 && counts["OUTCOME"] == 1 && counts["DRIVER"] == 2 && counts["GUARDRAIL"] == 3 && counts["FOUNDATION"] == 3 && counts["COHERENCE"] == 2 && counts["REGRESSION"] == 1
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
