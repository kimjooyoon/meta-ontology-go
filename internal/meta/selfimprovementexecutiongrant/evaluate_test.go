package selfimprovementexecutiongrant

import (
	"os"
	"testing"
)

func testProgram(t *testing.T) PolicyProgram {
	t.Helper()
	program, err := CompilePolicy(os.DirFS("../../.."), PolicyPath)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func TestPolicyUsesFirstClassSemanticIR(t *testing.T) {
	program := testProgram(t)
	if program.Evidence.StateCount != 12 || program.Evidence.TransitionCount != 9 || program.Evidence.CaseCount != 9 || program.Evidence.ClosedCases != 3 || program.Evidence.UnknownCases != 3 || program.Evidence.RefutedCases != 3 {
		t.Fatalf("execution grant policy denominator drifted: %#v", program.Evidence)
	}
	if program.Evidence.SourceDigest == "" || program.Evidence.CanonicalDigest == "" || program.Evidence.SemanticIRDigest == "" || !program.Inventory.Observed {
		t.Fatal("execution grant policy did not retain parser, formatter, semantic IR, and inventory evidence")
	}
}

func TestCanonicalCasesAreNineSeparateGrantCases(t *testing.T) {
	report, err := BuildCanonicalCaseReport(testProgram(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCanonicalCases(report); err != nil {
		t.Fatal(err)
	}
	if report.StructuralSeparateGrantEdgesBefore != 0 || report.StructuralSeparateGrantEdgesAfter != 1 || report.CanonicalGrantedCases != 2 || report.CanonicalExecutionCount != 0 || report.GrantConsumedUses != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.FallbackAccepted != 0 {
		t.Fatalf("canonical grant boundary drifted: %#v", report)
	}
}

func TestExplicitAllowProducesUnconsumedGrant(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	input := GrantInput{Request: request, DecisionInputs: []GrantDecisionInput{fixtureDecision(request, DecisionAllow)}}
	resolution := Evaluate(program, input)
	if resolution.Decision != DecisionClosed || resolution.Resolution != ResolutionGrantedUnconsumed || !resolution.GrantAllowsExecution || resolution.RemainingUses != 1 || resolution.ConsumedUses != 0 || resolution.ExecutionCount != 0 || resolution.OneUseEnforced {
		t.Fatalf("ALLOW crossed the execution or consumption boundary: %#v", resolution)
	}
	if err := ValidateGrantReceipt(*resolution.Receipt); err != nil {
		t.Fatal(err)
	}
	verification := Verify(program, input, resolution)
	if !verification.Verified || verification.ExecutionCount != 0 || verification.GrantConsumedUses != 0 {
		t.Fatalf("independent grant replay failed: %#v", verification)
	}
}

func TestExplicitDenyClosesWithoutGrant(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	resolution := Evaluate(program, GrantInput{Request: request, DecisionInputs: []GrantDecisionInput{fixtureDecision(request, DecisionDeny)}})
	if resolution.Decision != DecisionClosed || resolution.Resolution != ResolutionDenied || resolution.GrantAllowsExecution || resolution.RemainingUses != 0 || resolution.ExecutionCount != 0 {
		t.Fatalf("DENY did not close without execution authority: %#v", resolution)
	}
	if err := VerifyGrantResolution(resolution); err != nil {
		t.Fatal(err)
	}
}

func TestLiveNoDecisionIsSixFieldUnknown(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	resolution := Evaluate(program, GrantInput{Request: request, Live: true})
	if resolution.Decision != DecisionUnknown || resolution.Resolution != ResolutionLower || resolution.Reason != ReasonMissingDecision || resolution.Unknown == nil || resolution.Unknown.BlockedBy != "explicit_execution_grant_decision" {
		t.Fatalf("live request did not remain blocked: %#v", resolution)
	}
	if resolution.Metrics.LiveGrantRequests != 1 || resolution.Metrics.LiveGrants != 0 || resolution.ExecutionCount != 0 || resolution.ConsumedUses != 0 {
		t.Fatalf("live grant metrics crossed the boundary: %#v", resolution.Metrics)
	}
	if verification := Verify(program, GrantInput{Request: request, Live: true}, resolution); !verification.Verified {
		t.Fatalf("independent UNKNOWN replay failed: %#v", verification)
	}
}

func TestScopeDuplicateAndSafetyContradictionsRefute(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	scope := request
	scope.V25.BoundedTarget = "UNBOUNDED"
	scope.Digest = requestDigest(scope)
	if resolution := Evaluate(program, GrantInput{Request: scope, DecisionInputs: []GrantDecisionInput{fixtureDecision(scope, DecisionAllow)}}); resolution.Decision != DecisionRefuted || resolution.Reason != ReasonScopeMismatch {
		t.Fatalf("scope contradiction was not refuted: %#v", resolution)
	}
	allow := fixtureDecision(request, DecisionAllow)
	deny := fixtureDecision(request, DecisionDeny)
	if resolution := Evaluate(program, GrantInput{Request: request, DecisionInputs: []GrantDecisionInput{allow, deny}}); resolution.Decision != DecisionRefuted || resolution.Reason != ReasonDuplicate {
		t.Fatalf("conflicting duplicate was not refuted: %#v", resolution)
	}
	unsafe := request
	unsafe.V25.MaxExecutions = 2
	unsafe.Digest = requestDigest(unsafe)
	if resolution := Evaluate(program, GrantInput{Request: unsafe, DecisionInputs: []GrantDecisionInput{fixtureDecision(unsafe, DecisionAllow)}}); resolution.Decision != DecisionRefuted || resolution.Reason != ReasonUnsafe {
		t.Fatalf("unsafe grant was not refuted: %#v", resolution)
	}
}
