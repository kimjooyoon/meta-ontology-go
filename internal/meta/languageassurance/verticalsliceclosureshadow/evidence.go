package verticalsliceclosureshadow

import (
	_ "embed"
	"encoding/json"
)

//go:embed evidence/assurance.json
var embeddedAssurance []byte

//go:embed evidence/denominator.json
var embeddedDenominator []byte

func EmbeddedAssurance() []byte {
	return append([]byte(nil), embeddedAssurance...)
}

func EmbeddedDenominator() []byte {
	return append([]byte(nil), embeddedDenominator...)
}

type assuranceSummary struct {
	DenominatorTotal     int `json:"denominator_total"`
	Operating            int `json:"operating"`
	NotImplemented       int `json:"not_implemented"`
	CoverageBPS          int `json:"implementation_coverage_bps"`
	UnknownTopDecisions  int `json:"unknown_top_decisions"`
	UnresolvedIndicators int `json:"unresolved_indicators"`
	ViolatedGuardrails   int `json:"violated_guardrails"`
	RepositoryWrites     int `json:"repository_writes"`
}

type assuranceObligation struct {
	MetricID      string `json:"metric_id"`
	Status        string `json:"status"`
	Resolution    string `json:"resolution"`
	MetaOperation string `json:"meta_operation"`
}

type assuranceReport struct {
	Schema            string                `json:"schema"`
	SubjectSHA        string                `json:"subject_sha"`
	AssuranceDecision string                `json:"assurance_decision"`
	CandidateDecision string                `json:"candidate_decision"`
	Summary           assuranceSummary      `json:"summary"`
	Obligations       []assuranceObligation `json:"obligations"`
}

func inspectAssurance(raw []byte) (string, string, string) {
	if len(raw) == 0 {
		return "", ResolutionLower, ReasonAssuranceMissing
	}
	if digestBytes(raw) != AssuranceDigest {
		return "", ResolutionInvariant, ReasonAssuranceDigest
	}
	var report assuranceReport
	if json.Unmarshal(raw, &report) != nil || !validAssurance(report) {
		return "", ResolutionInvariant, ReasonAssuranceBase
	}
	return report.SubjectSHA, ResolutionExact, ""
}

func validAssurance(report assuranceReport) bool {
	summary := report.Summary
	return report.Schema == "gooo/language-assurance-report/v1" &&
		report.SubjectSHA == PredecessorSHA && report.AssuranceDecision == "PARTIAL" &&
		report.CandidateDecision == "ALLOW_LIMITED" && summary.DenominatorTotal == officialTotal &&
		summary.Operating == beforeOperating && summary.NotImplemented == 2 &&
		summary.CoverageBPS == beforeCoverageBPS && summary.UnknownTopDecisions == 0 &&
		summary.UnresolvedIndicators == 0 && summary.ViolatedGuardrails == 0 &&
		summary.RepositoryWrites == 0 && validTargetObligation(report.Obligations)
}

func validTargetObligation(obligations []assuranceObligation) bool {
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
