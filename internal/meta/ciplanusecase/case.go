package ciplanusecase

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"
)

func evaluateCase(spec CaseSpec, input Input) (CaseResult, bool) {
	result := CaseResult{
		ID: spec.ID, ExpectedDecision: spec.ExpectedDecision, ProofChoice: spec.ProofChoice,
		Status: "UNSATISFIED", Unknowns: []metainvocation.UnknownCause{}, ClaimStatuses: map[string]string{},
	}
	report, exists := input.Reports[spec.ID]
	if !exists {
		result.ObservedDecision = "MISSING"
		result.EvidenceDigest = digest(result)
		return result, false
	}
	result.ObservedDecision = report.Decision
	result.Unknowns = append(result.Unknowns, report.Unknowns...)
	for _, claim := range report.Claims {
		result.ClaimStatuses[claim.ID] = claim.Status
	}
	replay, replayExists := input.Replays[spec.ID]
	satisfied := metainvocation.Validate(report) == nil && replayExists && replay.ReportDigest == report.ReportDigest &&
		report.CaseID == spec.ID && report.Decision == spec.ExpectedDecision && exactClaimTransitions(report)
	if spec.ExpectedDecision == metainvocation.DecisionPass {
		golden, goldenExists := input.Goldens[spec.ID]
		satisfied = satisfied && goldenExists && reflect.DeepEqual(ProjectGolden(report), golden) && len(report.Unknowns) == 0
	} else if spec.ExpectedDecision == metainvocation.DecisionUnknown {
		satisfied = satisfied && report.Resolution == metainvocation.ResolutionLower && len(report.Unknowns) == 1 && exactUnknown(report.Unknowns[0])
	} else {
		satisfied = satisfied && report.Resolution == metainvocation.ResolutionExact && len(report.Unknowns) == 0
	}
	if satisfied {
		result.Status = "SATISFIED"
	}
	result.EvidenceDigest = caseEvidenceDigest(report, replay, input.Goldens[spec.ID])
	return result, satisfied
}

func caseEvidenceDigest(report, replay metainvocation.Report, golden GoldenPlan) string {
	return digest(struct {
		Report string     `json:"report"`
		Replay string     `json:"replay"`
		Golden GoldenPlan `json:"golden"`
	}{Report: report.ReportDigest, Replay: replay.ReportDigest, Golden: golden})
}
