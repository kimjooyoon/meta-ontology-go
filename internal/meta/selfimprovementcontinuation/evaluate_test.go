package selfimprovementcontinuation

import (
	"encoding/json"
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

func TestScalarIdentityParsingIsStrictAndRefuted(t *testing.T) {
	program := testProgram(t)
	encode := func(input ContinuationInput, runID, attempt, artifactID string) []byte {
		t.Helper()
		raw, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		for key, value := range map[string]string{
			"source_run_id": runID, "source_run_attempt": attempt, "source_artifact_id": artifactID,
		} {
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			fields[key] = encoded
		}
		encoded, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}

	var valid ContinuationInput
	if err := json.Unmarshal(encode(canonicalInput(), "1", "1", "1"), &valid); err != nil {
		t.Fatal(err)
	}
	if valid.SourceRunID != 1 || valid.SourceRunAttempt != 1 || valid.SourceArtifactID != 1 || len(valid.ParseErrors) != 0 {
		t.Fatalf("valid scalar identity was not strictly parsed: %#v", valid)
	}
	if resolution := Evaluate(program, valid); resolution.Decision != DecisionClosed || resolution.Reason != ReasonExact {
		t.Fatalf("valid scalar identity was not accepted exactly: %#v", resolution)
	}

	var malformed ContinuationInput
	if err := json.Unmarshal(encode(canonicalInput(), "9223372036854775808", "0", "not-decimal"), &malformed); err != nil {
		t.Fatal(err)
	}
	resolution := Evaluate(program, malformed)
	if resolution.Decision != DecisionRefuted || resolution.Reason != ReasonMalformedIdentity || len(resolution.ContradictoryFields) != 3 {
		t.Fatalf("malformed scalar identity was not REFUTED: %#v", resolution)
	}
	report := BuildReport(program, malformed)
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip Report
	if err := json.Unmarshal(raw, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if err := ValidateReport(roundTrip); err != nil {
		t.Fatalf("REFUTED scalar identity report did not replay: %v", err)
	}
}
