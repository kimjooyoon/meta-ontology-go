package selfimprovementcontinuation

import (
	"errors"
	"fmt"
)

func canonicalInput() ContinuationInput {
	sha := "0123456789abcdef0123456789abcdef01234567"
	digest := digestBytes([]byte("continuation-artifact"))
	return ContinuationInput{
		SourceWorkflowName: SourceWorkflowName, SourceWorkflowPath: SourceWorkflowPath, SourceRepository: SourceRepository,
		SourceEvent: SourceEvent, SourceRef: SourceRef, SourceHeadSHA: sha, SourceRunID: 1, SourceRunAttempt: 1,
		SourceArtifactName: "self-improvement-candidate-authorization-" + sha, SourceArtifactID: 7,
		SourceArtifactArchiveDigest: digest, SourceArtifactObservedDigest: digest, SourceReceiptDigest: digestBytes([]byte("continuation-receipt")),
		TargetWorkflowName: V25WorkflowName, TargetWorkflowPath: V25WorkflowPath, DispatchRef: DispatchRef, DispatchMode: DispatchMode,
	}
}

func BuildCanonicalCaseReport(program PolicyProgram) (CanonicalCaseReport, error) {
	base := canonicalInput()
	cases := []struct {
		id       string
		input    func(ContinuationInput) ContinuationInput
		expected Decision
		reason   string
	}{
		{"exact-auth-to-v25", func(input ContinuationInput) ContinuationInput { return input }, DecisionClosed, ReasonExact},
		{"deterministic-replay", func(input ContinuationInput) ContinuationInput { input.Replay = true; return input }, DecisionClosed, ReasonReplay},
		{"idempotent-duplicate-dispatch-v25-to-v26", func(input ContinuationInput) ContinuationInput {
			input.SourceWorkflowName = V25WorkflowName
			input.SourceWorkflowPath = V25WorkflowPath
			input.SourceEvent = DispatchMode
			input.SourceArtifactName = "self-improvement-execution-contract-" + input.SourceHeadSHA
			input.TargetWorkflowName = V26WorkflowName
			input.TargetWorkflowPath = V26WorkflowPath
			input.DuplicateDispatch = true
			return input
		}, DecisionClosed, ReasonIdempotent},
		{"missing-dispatch-identity", func(input ContinuationInput) ContinuationInput { input.SourceRunID = 0; return input }, DecisionUnknown, ReasonMissingIdentity},
		{"missing-artifact-receipt", func(input ContinuationInput) ContinuationInput { input.SourceReceiptDigest = ""; return input }, DecisionUnknown, ReasonMissingReceipt},
		{"missing-head-lineage", func(input ContinuationInput) ContinuationInput { input.SourceHeadSHA = ""; return input }, DecisionUnknown, ReasonMissingLineage},
		{"workflow-identity-mismatch", func(input ContinuationInput) ContinuationInput {
			input.SourceWorkflowPath = "wrong/workflow.yml"
			return input
		}, DecisionRefuted, ReasonWorkflowMismatch},
		{"artifact-digest-mismatch", func(input ContinuationInput) ContinuationInput {
			input.SourceArtifactObservedDigest = digestBytes([]byte("tampered"))
			return input
		}, DecisionRefuted, ReasonDigestMismatch},
		{"unauthorized-dispatch", func(input ContinuationInput) ContinuationInput { input.ExecutionAuthorized = true; return input }, DecisionRefuted, ReasonUnauthorized},
	}
	report := CanonicalCaseReport{Schema: CanonicalCasesSchema, Policy: program.Evidence, CaseDenominator: 9, Counts: map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}, Cases: []CanonicalCase{}, ReplayEqual: true, Decision: DecisionClosed, Resolution: ResolutionExact, Reason: "NINE_CANONICAL_CI_CONTINUATION_CASES", Metrics: Metrics{WorkflowRunContinuationEdgesBefore: DepthEdgesBefore, WorkflowRunContinuationEdgesAfter: DepthEdgesAfter, ExactIdentityBindingsBefore: IdentityBindingsBefore, ExactIdentityBindingsAfter: IdentityBindingsAfter, ManualDispatchesBefore: 0, ManualDispatchesAfter: 0, LiveGrantDecision: 0, LiveGrants: 0, LiveExecutionCount: 0, GrantConsumedUses: 0, RepositoryWrites: 0, LocalTestExecutions: 0, CanonicalExecutionCount: 0, CanonicalCases: 9, SixFieldUnknowns: 0, UnauthorizedDispatches: 0, FallbackAccepted: 0, IndependentReplayComparisons: 1, ArtifactFiles: ArtifactFiles, ArtifactTypes: ArtifactTypes, GoPhysicalLines: program.GoPhysicalLines, GoooPhysicalLines: program.GoooPhysicalLines, PerformanceImprovement: PerformanceUnknown, CounterexampleRunID: CounterexampleRunID}}
	for _, spec := range cases {
		input := spec.input(base)
		first := Evaluate(program, input)
		second := Evaluate(program, input)
		verification := Verify(program, BuildRequest(program, input), first)
		replayEqual := first.Digest == second.Digest
		if !replayEqual {
			report.ReplayEqual = false
		}
		pass := first.Decision == spec.expected && first.Reason == spec.reason && replayEqual && verification.Verified
		if !pass {
			return CanonicalCaseReport{}, fmt.Errorf("canonical continuation case %s failed", spec.id)
		}
		report.Cases = append(report.Cases, CanonicalCase{ID: spec.id, ExpectedDecision: spec.expected, ExpectedReason: spec.reason, ActualDecision: first.Decision, ActualReason: first.Reason, Unknown: first.Unknown, Pass: pass})
		report.Counts[string(first.Decision)]++
		if first.Unknown != nil {
			report.Metrics.SixFieldUnknowns++
		}
	}
	report.ClosedCases, report.UnknownCases, report.RefutedCases = report.Counts["CLOSED"], report.Counts["UNKNOWN"], report.Counts["REFUTED"]
	report.Digest = canonicalDigest(report)
	return report, nil
}

func ValidateCanonicalCases(report CanonicalCaseReport) error {
	if report.Schema != CanonicalCasesSchema || report.CaseDenominator != 9 || len(report.Cases) != 9 || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || report.Counts["CLOSED"] != 3 || report.Counts["UNKNOWN"] != 3 || report.Counts["REFUTED"] != 3 || !report.ReplayEqual || report.Metrics.WorkflowRunContinuationEdgesBefore != DepthEdgesBefore || report.Metrics.WorkflowRunContinuationEdgesAfter != DepthEdgesAfter || report.Metrics.ExactIdentityBindingsBefore != IdentityBindingsBefore || report.Metrics.ExactIdentityBindingsAfter != IdentityBindingsAfter || report.Metrics.ManualDispatchesBefore != 0 || report.Metrics.ManualDispatchesAfter != 0 || report.Metrics.LiveGrantDecision != 0 || report.Metrics.LiveGrants != 0 || report.Metrics.LiveExecutionCount != 0 || report.Metrics.GrantConsumedUses != 0 || report.Metrics.RepositoryWrites != 0 || report.Metrics.LocalTestExecutions != 0 || report.Metrics.CanonicalExecutionCount != 0 || report.Metrics.CanonicalCases != 9 || report.Metrics.SixFieldUnknowns != 3 || report.Metrics.UnauthorizedDispatches != 0 || report.Metrics.FallbackAccepted != 0 || report.Metrics.IndependentReplayComparisons != 1 || report.Metrics.ArtifactFiles != ArtifactFiles || report.Metrics.ArtifactTypes != ArtifactTypes || report.Metrics.PerformanceImprovement != PerformanceUnknown || report.Metrics.CounterexampleRunID != CounterexampleRunID || report.Digest != canonicalDigest(report) {
		return errors.New("canonical continuation cases are not exact")
	}
	for _, item := range report.Cases {
		if !item.Pass || item.ExpectedDecision != item.ActualDecision || item.ExpectedReason != item.ActualReason {
			return errors.New("canonical continuation case failed")
		}
		if item.ActualDecision == DecisionUnknown && item.Unknown == nil {
			return errors.New("canonical UNKNOWN continuation lacks six fields")
		}
		if item.ActualDecision != DecisionUnknown && item.Unknown != nil {
			return errors.New("canonical non-UNKNOWN continuation contains unknown evidence")
		}
	}
	return nil
}
