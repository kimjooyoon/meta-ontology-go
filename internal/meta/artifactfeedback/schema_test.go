package artifactfeedback

import "testing"

func TestCanonicalProgramIsMetaBound(t *testing.T) {
	program := CanonicalProgram()
	if err := Validate(program); err != nil {
		t.Fatal(err)
	}
	if len(program.MetaOperations) != 6 || len(program.Indicators) != 7 {
		t.Fatalf("program members = %d/%d, want 6/7", len(program.MetaOperations), len(program.Indicators))
	}
}

func TestDuplicateFeedbackIndicatorFailsClosed(t *testing.T) {
	program := CanonicalProgram()
	program.Indicators = append(program.Indicators, program.Indicators[0])
	if err := Validate(program); err == nil {
		t.Fatal("duplicate feedback indicator was accepted")
	}
}

func TestMissingRegressionProofFailsClosed(t *testing.T) {
	program := CanonicalProgram()
	kept := program.MetaOperations[:0]
	for _, operation := range program.MetaOperations {
		if operation.ProofChoice != ProofRegression {
			kept = append(kept, operation)
		}
	}
	program.MetaOperations = kept
	if err := Validate(program); err == nil {
		t.Fatal("missing regression proof was accepted")
	}
}
