package integrationprogress

import "testing"

func TestMergeBeforeEvidenceFailsClosed(t *testing.T) {
	input := completeObservation()
	pull := fixturePull(&input, 550)
	pull.AuthoritativeRun.Artifact.CreatedAt = "2026-08-01T12:05:00Z"
	report := Evaluate(input, true)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Summary.RefutedCells == 0 {
		t.Fatalf("contradiction report = %#v", report.Summary)
	}
}

func TestGeneratedMetaProgramIsDeterministic(t *testing.T) {
	first, second := string(RenderProgram()), string(RenderProgram())
	if first != second || first == "" {
		t.Fatal("generated meta program replay differs")
	}
}
