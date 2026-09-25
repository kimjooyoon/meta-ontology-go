package governancesnapshot

import (
	"fmt"
	"strings"
)

func finishReport(report *Report, contract Contract) {
	report.Summary = summarize(report.Cells)
	report.Unknowns = collectUnknowns(report.Cells)
	report.Decision, report.Resolution, report.Reason = terminal(report)
	report.SettingsHealthy = report.Decision == DecisionClosed
	report.Authority = AuthorityCounters{RepositoryWrites: report.RepositoryWrites, BranchSettingWrites: report.BranchSettingWrites,
		LocalTestExecutions: report.LocalTestExecutions, CrossProjectGates: report.CrossProjectGates}
	report.HumanReport = humanReport(*report, contract)
	report.ReportDigest = sealReport(*report)
}

func summarize(cells []CellObservation) Summary {
	result := Summary{CellsTotal: len(cells)}
	for _, cell := range cells {
		switch cell.Decision {
		case DecisionClosed:
			result.ClosedCells++
		case DecisionUnknown:
			result.UnknownCells++
		case DecisionRefuted:
			result.RefutedCells++
		}
		switch cell.ProofChoice {
		case "FOUNDATION":
			result.FoundationCells++
		case "COHERENCE":
			result.CoherenceCells++
		case "REGRESSION":
			result.RegressionCells++
		}
		switch cell.Indicator {
		case "DRIVER":
			result.DriverCells++
		case "OUTCOME":
			result.OutcomeCells++
		case "GUARDRAIL":
			result.GuardrailCells++
		}
	}
	return result
}

func collectUnknowns(cells []CellObservation) []Unknown {
	result := []Unknown{}
	for _, cell := range cells {
		if cell.Unknown != nil {
			result = append(result, *cell.Unknown)
		}
	}
	return result
}

func terminal(report *Report) (string, string, string) {
	if report.DefaultBranch == "" {
		if len(report.Unknowns) > 0 {
			return DecisionUnknown, ResolutionLower, "PUBLIC_SNAPSHOT_MISSING"
		}
	}
	for _, cell := range report.Cells {
		if cell.Decision == DecisionRefuted {
			if cell.ID == "DEV_STATUS_ENFORCEMENT" || cell.ID == "DEV_CONTEXT_SET" {
				return DecisionRefuted, ResolutionExact, "DEV_LIVE_PROTECTION_DRIFT"
			}
			if report.DefaultBranch == "" {
				return DecisionRefuted, ResolutionExact, "MALFORMED_PUBLIC_PAYLOAD"
			}
			return DecisionRefuted, ResolutionExact, cell.Reason
		}
	}
	if len(report.Unknowns) > 0 {
		if allRequestsUnavailable(report.Source.Requests) {
			return DecisionUnknown, ResolutionLower, "PUBLIC_SNAPSHOT_MISSING"
		}
		return DecisionUnknown, ResolutionLower, "GOVERNANCE_DEPENDENCY_UNKNOWN"
	}
	return DecisionClosed, ResolutionExact, "GOVERNANCE_SNAPSHOT_CONFORMANT"
}

func allRequestsUnavailable(requests []RequestObservation) bool {
	if len(requests) == 0 {
		return true
	}
	for _, request := range requests {
		if request.State == "PRESENT" {
			return false
		}
	}
	return true
}

func humanReport(report Report, contract Contract) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Governance snapshot: %s/%s (%s)\n", report.Decision, report.Resolution, report.Reason)
	fmt.Fprintf(&b, "Repository=%s default_branch=%s head=%s\n", report.Repository, report.DefaultBranch, report.HeadSHA)
	fmt.Fprintf(&b, "Cells=%d closed=%d unknown=%d refuted=%d; proof FOUNDATION/COHERENCE/REGRESSION=%d/%d/%d\n",
		report.Summary.CellsTotal, report.Summary.ClosedCells, report.Summary.UnknownCells, report.Summary.RefutedCells,
		report.Summary.FoundationCells, report.Summary.CoherenceCells, report.Summary.RegressionCells)
	fmt.Fprintf(&b, "Indicators DRIVER/OUTCOME/GUARDRAIL=%d/%d/%d\n", report.Summary.DriverCells, report.Summary.OutcomeCells, report.Summary.GuardrailCells)
	for _, branch := range report.Branches {
		fmt.Fprintf(&b, "%s protected=%t enforcement=%s contexts=%d\n", branch.Branch, branch.Protected, branch.Enforcement, len(branch.Contexts))
	}
	fmt.Fprintf(&b, "Rulesets=%d disabled_authority_rejected=true\n", len(report.Rulesets))
	fmt.Fprintf(&b, "Requests=%d normalized_payloads=%d payload_digest_model=%s graph_activities=%d bindings=%d\n", len(report.Source.Requests), len(report.Source.Payloads), report.Source.PayloadDigestModel, report.Graph.ActivityCount, report.Graph.BindingCount)
	fmt.Fprintf(&b, "repository_writes=%d branch_setting_writes=%d local_test_executions=%d cross_project_required_gates=%d improvement=%s\n",
		report.RepositoryWrites, report.BranchSettingWrites, report.LocalTestExecutions, report.CrossProjectGates, report.Improvement)
	for _, item := range report.Cases {
		fmt.Fprintf(&b, "case %s: %s/%s %s\n", item.ID, item.Decision, item.Resolution, item.Reason)
	}
	for _, unknown := range report.Unknowns {
		fmt.Fprintf(&b, "unknown stage=%s step=%s reason=%s class=%s next=%s blocked_by=%v\n",
			unknown.Stage, unknown.Step, unknown.Reason, unknown.UnknownClass, unknown.NextOperation, unknown.BlockedBy)
	}
	return b.String()
}

func sealReport(report Report) string {
	report.ReportDigest = ""
	return digestJSON(report)
}
