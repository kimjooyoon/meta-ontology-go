package ciplanusecase

import (
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"
)

func Evaluate(input Input) Report {
	summary := summarize(input)
	cases := make([]CaseResult, 0, len(input.Contract.Cases))
	for _, spec := range input.Contract.Cases {
		result, satisfied := evaluateCase(spec, input)
		cases = append(cases, result)
		if satisfied {
			summary.CasesSatisfied++
		}
	}
	indicators := buildIndicators(summary, input.Contract.Limits)
	proofs := buildProofs(summary, input.Contract)
	decision := "PASS"
	for _, indicator := range indicators {
		if indicator.Status != "SATISFIED" {
			decision = "FAIL_CLOSED"
		}
	}
	for _, proof := range proofs {
		if proof.Status != "SATISFIED" {
			decision = "FAIL_CLOSED"
		}
	}
	report := Report{
		Schema: ReportSchema, Decision: decision, Resolution: "EXACT",
		Interpretation: "MINIMAL_CI_PLAN_VALUE_OBSERVED", ContractDigest: digest(input.Contract),
		Cases: cases, Summary: summary, Indicators: indicators, ReaderViews: buildReaderViews(indicators),
		Proofs: proofs, NotClaimed: append([]string(nil), input.Contract.NotClaimed...),
	}
	return seal(report)
}

func summarize(input Input) Summary {
	summary := Summary{GoooFiles: input.Source.GoooFiles, GoFiles: input.Source.GoFiles, GoooLines: input.Source.GoooLines, GoLines: input.Source.GoLines}
	if input.GeneratedReplay {
		summary.GeneratedReplays = 1
	}
	for _, spec := range input.Contract.Cases {
		report, ok := input.Reports[spec.ID]
		if !ok {
			continue
		}
		switch report.Decision {
		case metainvocation.DecisionPass:
			summary.PassDecisions++
		case metainvocation.DecisionClosed:
			summary.FailClosedDecisions++
		case metainvocation.DecisionUnknown:
			summary.UnknownDecisions++
		}
		if replay, exists := input.Replays[spec.ID]; exists && replay.ReportDigest == report.ReportDigest {
			summary.DeterministicReplays++
		}
		if golden, exists := input.Goldens[spec.ID]; exists && reflect.DeepEqual(ProjectGolden(report), golden) {
			summary.GoldenPlans++
		}
		for _, check := range report.Plan.Checks {
			summary.RuleEvidenceRefs += len(check.Reasons)
		}
		summary.PersistentClaims += len(report.Claims)
		for _, claim := range report.Claims {
			if claim.Reason == "UNKNOWN_AT_RULE_SELECTION" {
				summary.DirectUnknownClaims++
			}
			if claim.Reason == "DEPENDENCY_BLOCKED" {
				summary.DependencyBlocked++
			}
			if claim.Status == metainvocation.ClaimRefuted {
				summary.RefutedClaims++
			}
		}
		summary.RepositoryWrites += report.Effects.RepositoryWrites
		if report.Effects.MutationAuthority {
			summary.MutationAuthority++
		}
	}
	profiles := map[string]ProfileSample{}
	for _, sample := range input.Profile.Samples {
		profiles[sample.CaseID] = sample
	}
	for _, spec := range input.Contract.Cases {
		sample, ok := profiles[spec.ID]
		if !ok {
			continue
		}
		summary.ResourceSamples++
		if sample.WallMS > summary.MaxWallMS {
			summary.MaxWallMS = sample.WallMS
		}
		if sample.PeakRSSKiB > summary.MaxPeakRSSKiB {
			summary.MaxPeakRSSKiB = sample.PeakRSSKiB
		}
		if sample.ReceiptBytes > summary.MaxReceiptBytes {
			summary.MaxReceiptBytes = sample.ReceiptBytes
		}
	}
	if len(input.Profile.Samples) != input.Contract.Denominator {
		summary.ResourceSamples = -1
	}
	return summary
}

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
	result.EvidenceDigest = digest(struct {
		Report string     `json:"report"`
		Replay string     `json:"replay"`
		Golden GoldenPlan `json:"golden"`
	}{Report: report.ReportDigest, Replay: replay.ReportDigest, Golden: input.Goldens[spec.ID]})
	return result, satisfied
}

func exactClaimTransitions(report metainvocation.Report) bool {
	if len(report.Claims) != 3 || report.Claims[0].ID != "source-program-binding" || report.Claims[0].Status != metainvocation.ClaimDischarged {
		return false
	}
	ruleClaim, planClaim := report.Claims[1], report.Claims[2]
	switch report.Decision {
	case metainvocation.DecisionPass:
		return ruleClaim.Status == metainvocation.ClaimDischarged && planClaim.Status == metainvocation.ClaimDischarged
	case metainvocation.DecisionClosed:
		return ruleClaim.Status == metainvocation.ClaimRefuted && planClaim.Status == metainvocation.ClaimOpen && planClaim.Reason == "DEPENDENCY_BLOCKED"
	case metainvocation.DecisionUnknown:
		return ruleClaim.Status == metainvocation.ClaimOpen && ruleClaim.Reason == "UNKNOWN_AT_RULE_SELECTION" && planClaim.Status == metainvocation.ClaimOpen && planClaim.Reason == "DEPENDENCY_BLOCKED"
	default:
		return false
	}
}

func exactUnknown(cause metainvocation.UnknownCause) bool {
	return cause.Stage == "RULE_SELECTION" && cause.Step == "classify-change" && cause.Reason == "NO_REGISTERED_RULE" && cause.File != ""
}

func ProjectGolden(report metainvocation.Report) GoldenPlan {
	golden := GoldenPlan{Schema: "gooo/ci-plan-golden/v1", CaseID: report.CaseID, Checks: make([]GoldenCheck, 0, len(report.Plan.Checks))}
	for _, check := range report.Plan.Checks {
		projected := GoldenCheck{ID: check.ID, Command: check.Command, Files: append([]string(nil), check.Files...), Reasons: make([]GoldenReason, 0, len(check.Reasons))}
		for _, reason := range check.Reasons {
			projected.Reasons = append(projected.Reasons, GoldenReason{Operation: reason.Operation, File: reason.File, SourcePath: reason.Source.Path, SourceLine: reason.Source.StartLine})
		}
		golden.Checks = append(golden.Checks, projected)
	}
	return golden
}

func buildProofs(summary Summary, contract Contract) []Proof {
	foundation := status(contract.Schema == ContractSchema && contract.Denominator == 12 && summary.ResourceSamples == 12)
	coherence := status(summary.CasesSatisfied == 12 && summary.DeterministicReplays == 12 && summary.GoldenPlans == 4 && summary.GeneratedReplays == 1)
	regression := status(summary.PassDecisions == 4 && summary.FailClosedDecisions == 4 && summary.UnknownDecisions == 4 && summary.RepositoryWrites == 0 && summary.MutationAuthority == 0)
	return []Proof{
		{Choice: "FOUNDATION", Status: foundation, Evidence: []string{fmt.Sprintf("contract=%d resource-samples=%d", contract.Denominator, summary.ResourceSamples)}},
		{Choice: "COHERENCE", Status: coherence, Evidence: []string{fmt.Sprintf("cases=%d replays=%d golden=%d generated=%d", summary.CasesSatisfied, summary.DeterministicReplays, summary.GoldenPlans, summary.GeneratedReplays)}},
		{Choice: "REGRESSION", Status: regression, Evidence: []string{fmt.Sprintf("pass=%d fail=%d unknown=%d writes=%d authority=%d", summary.PassDecisions, summary.FailClosedDecisions, summary.UnknownDecisions, summary.RepositoryWrites, summary.MutationAuthority)}},
	}
}

func status(value bool) string {
	if value {
		return "SATISFIED"
	}
	return "UNSATISFIED"
}
