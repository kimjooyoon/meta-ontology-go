package semanticresolution

import "testing"

func TestCanonicalProgramValidates(t *testing.T) {
	if err := Validate(CanonicalProgram()); err != nil {
		t.Fatal(err)
	}
}

func TestMissingRegressionProofFailsClosed(t *testing.T) {
	program := CanonicalProgram()
	program.MetaOperations[3].ProofChoice = ProofCoherence
	if err := Validate(program); err == nil {
		t.Fatal("missing regression proof was accepted")
	}
}
