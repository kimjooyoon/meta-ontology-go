package selfimprovementcontinuation

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
		t.Fatalf("continuation policy denominator drifted: %#v", program.Evidence)
	}
	if program.Evidence.SourceDigest == "" || program.Evidence.CanonicalDigest == "" || program.Evidence.SemanticIRDigest == "" {
		t.Fatal("continuation policy lost parser, formatter, or semantic IR evidence")
	}
}

func TestCanonicalContinuationCases(t *testing.T) {
	report, err := BuildCanonicalCaseReport(testProgram(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCanonicalCases(report); err != nil {
		t.Fatal(err)
	}
}

func TestMissingIdentityIsSixFieldUnknown(t *testing.T) {
	program := testProgram(t)
	input := canonicalInput()
	input.SourceRunID = 0
	resolution := Evaluate(program, input)
	if resolution.Decision != DecisionUnknown || resolution.Reason != ReasonMissingIdentity || resolution.Unknown == nil || resolution.Unknown.Stage == "" || resolution.Unknown.Step == "" || resolution.Unknown.Reason == "" || resolution.Unknown.UnknownClass == "" || resolution.Unknown.NextOperation == "" || len(resolution.Unknown.BlockedBy) == 0 {
		t.Fatalf("missing identity did not produce six-field UNKNOWN: %#v", resolution)
	}
}

func TestContradictionDominatesMissingEvidence(t *testing.T) {
	program := testProgram(t)
	input := canonicalInput()
	input.SourceArtifactObservedDigest = digestBytes([]byte("tampered"))
	input.SourceReceiptDigest = ""
	resolution := Evaluate(program, input)
	if resolution.Decision != DecisionRefuted || resolution.Reason != ReasonDigestMismatch {
		t.Fatalf("digest contradiction did not dominate missing evidence: %#v", resolution)
	}
}

func TestContinuationIsSchedulingOnly(t *testing.T) {
	program := testProgram(t)
	input := canonicalInput()
	report := BuildReport(program, input)
	if !report.Verification.Verified || report.Resolution.ExecutionAuthorized || report.Resolution.ExecutionGrants != 0 || report.Resolution.LiveGrantDecision != 0 || report.Resolution.LiveExecutionCount != 0 || report.Resolution.GrantConsumedUses != 0 || report.Resolution.RepositoryWrites != 0 || report.Resolution.LocalTestExecutions != 0 {
		t.Fatalf("continuation crossed authority boundary: %#v", report)
	}
}
