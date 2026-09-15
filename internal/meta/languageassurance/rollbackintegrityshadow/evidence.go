package rollbackintegrityshadow

import (
	_ "embed"
	"encoding/json"
)

//go:embed evidence/assurance.json
var embeddedAssurance []byte

func EmbeddedAssurance() []byte {
	return append([]byte(nil), embeddedAssurance...)
}

func inspectAssurance(raw []byte) (assuranceReport, string, string) {
	var report assuranceReport
	if len(raw) == 0 {
		return report, ResolutionLower, ReasonUnavailable
	}
	if digestBytes(raw) != AssuranceDigest {
		return report, ResolutionInvariant, ReasonDigest
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		return assuranceReport{}, ResolutionInvariant, ReasonBaseline
	}
	if !validAssuranceHeader(report) || !validTarget(report.Obligations) {
		return assuranceReport{}, ResolutionInvariant, ReasonBaseline
	}
	return report, ResolutionExact, ""
}

func validAssuranceHeader(report assuranceReport) bool {
	summary := report.Summary
	return report.Schema == "gooo/language-assurance-report/v1" &&
		report.SubjectSHA == PredecessorSHA && report.AssuranceDecision == "PARTIAL" &&
		report.CandidateDecision == "ALLOW_LIMITED" && summary.DenominatorTotal == 12 &&
		summary.Operating == 9 && summary.NotImplemented == 3 && summary.CoverageBPS == 7500 &&
		summary.UnknownTopDecisions == 0 && summary.UnresolvedIndicators == 0 &&
		summary.ViolatedGuardrails == 0 && summary.RepositoryWrites == 0
}

func validTarget(obligations []assuranceObligation) bool {
	matches := 0
	for _, obligation := range obligations {
		if obligation.MetricID != MetricID {
			continue
		}
		matches++
		if obligation.Status != "NOT_IMPLEMENTED" || obligation.Resolution != "NONE" ||
			obligation.MetaOperation != "" {
			return false
		}
	}
	return matches == 1
}
