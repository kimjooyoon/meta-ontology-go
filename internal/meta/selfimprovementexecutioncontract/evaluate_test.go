package selfimprovementexecutioncontract

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
	if program.Evidence.CaseCount != 9 || program.Evidence.ClosedCases != 3 || program.Evidence.UnknownCases != 3 || program.Evidence.RefutedCases != 3 {
		t.Fatalf("policy denominator drifted: %#v", program.Evidence)
	}
	if program.Evidence.SourceDigest == "" || program.Evidence.CanonicalDigest == "" || program.Evidence.SemanticIRDigest == "" {
		t.Fatal("policy did not retain parser, formatter, and semantic IR evidence")
	}
}

func TestCanonicalCasesAreNineSeparateV25Cases(t *testing.T) {
	report := BuildCanonicalCaseReport(testProgram(t))
	if report.CaseDenominator != 9 || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || !report.ReplayEqual {
		t.Fatalf("canonical report drifted: %#v", report)
	}
	if report.LiveExecutionCount != 0 || report.CanonicalExecutionCount != 0 || report.ExecutionGrants != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.FallbackAccepted != 0 {
		t.Fatalf("execution boundary drifted: %#v", report)
	}
	for _, current := range report.Cases {
		if !current.Pass {
			t.Fatalf("canonical case failed: %#v", current)
		}
	}
}

func TestV24AllowDoesNotAuthorizeExecution(t *testing.T) {
	program := testProgram(t)
	input := canonicalInput()
	input.Authorization.Decision = "ALLOW"
	resolution := Evaluate(program, input)
	if resolution.Decision != DecisionClosed || resolution.Resolution != ResolutionDeclared || resolution.ExecutionAuthorized || !resolution.ExecutionGrantRequired || resolution.MaxExecutions != 1 || resolution.RepositoryWritesAllowed {
		t.Fatalf("ALLOW crossed execution boundary: %#v", resolution)
	}
	if err := ValidateResolution(resolution); err != nil {
		t.Fatal(err)
	}
	verification := Verify(program, input, resolution)
	if !verification.Verified || verification.IndependentReplayComparisons != 1 {
		t.Fatalf("independent replay failed: %#v", verification)
	}
}

func TestUnknownHasExactlySixCausalFields(t *testing.T) {
	program := testProgram(t)
	input := canonicalInput()
	input.Observation = ObservationEvidence{}
	resolution := Evaluate(program, input)
	if resolution.Decision != DecisionUnknown || resolution.Resolution != ResolutionLower || resolution.Reason != ReasonMissingObservation || resolution.Unknown == nil {
		t.Fatalf("missing observation did not lower resolution: %#v", resolution)
	}
	unknown := resolution.Unknown
	if unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.UnknownClass == "" || unknown.NextOperation == "" || len(unknown.BlockedBy) == 0 {
		t.Fatalf("unknown lacks six fields: %#v", unknown)
	}
}

func TestContradictionRefutesBeforeUnknown(t *testing.T) {
	program := testProgram(t)
	input := canonicalInput()
	input.Observation.CandidateInputDigest = digestBytes([]byte("different"))
	if resolution := Evaluate(program, input); resolution.Decision != DecisionRefuted || resolution.Reason != ReasonCandidateConflict {
		t.Fatalf("digest contradiction was not refuted: %#v", resolution)
	}
	input = canonicalInput()
	input.Registry.MaxExecutions = 2
	if resolution := Evaluate(program, input); resolution.Decision != DecisionRefuted || resolution.Reason != ReasonSafetyConflict {
		t.Fatalf("execution ceiling contradiction was not refuted: %#v", resolution)
	}
}
