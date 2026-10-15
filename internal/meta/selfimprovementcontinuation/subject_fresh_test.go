package selfimprovementcontinuation

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSeparateContinuationOriginsFresh1613(t *testing.T) {
	program := testProgram(t)
	for _, grant := range []bool{false, true} {
		input := canonicalInput()
		input.SubjectSHA = strings.Repeat("c", 40)
		if grant {
			input.SourceWorkflowName, input.SourceWorkflowPath, input.SourceEvent = V25WorkflowName, V25WorkflowPath, DispatchMode
			input.TargetWorkflowName, input.TargetWorkflowPath = V26WorkflowName, V26WorkflowPath
			input.SourceArtifactName = "self-improvement-execution-contract-" + input.SubjectSHA
		}
		raw, _ := json.Marshal(input)
		var decoded ContinuationInput
		if err := json.Unmarshal(raw, &decoded); err != nil || decoded.SubjectSHA != input.SubjectSHA || decoded.SourceHeadSHA != input.SourceHeadSHA {
			t.Fatal("origin round-trip", err)
		}
		report := BuildReport(program, decoded)
		if err := ValidateReport(report); err != nil || report.Resolution.Decision != DecisionClosed || !report.Verification.Verified {
			t.Fatalf("separate origins: %+v %v", report.Resolution, err)
		}
		if grant {
			decoded.SourceArtifactName = "self-improvement-execution-contract-" + decoded.SourceHeadSHA
		} else {
			decoded.SourceArtifactName = "self-improvement-candidate-authorization-" + decoded.SubjectSHA
		}
		if got := Evaluate(program, decoded); got.Decision != DecisionRefuted {
			t.Fatal("swapped origin accepted", got)
		}
	}
}

func TestMalformedContinuationSubjectFresh1613(t *testing.T) {
	program, input := testProgram(t), canonicalInput()
	input.SubjectSHA = "not-a-commit"
	report := BuildReport(program, input)
	if report.Resolution.Decision != DecisionRefuted || !report.Verification.Verified {
		t.Fatal("malformed subject accepted")
	}
}
