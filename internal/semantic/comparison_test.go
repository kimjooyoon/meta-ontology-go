package semantic

import "testing"

func TestCompareIRSeparatesEquivalenceFromExactHostEvidence(t *testing.T) {
	left := selfHostingIR(t, GoHostedCompilerID)
	right := selfHostingIR(t, GoooHostedCompilerID)
	comparison := CompareIR(left, right)
	if !comparison.LeftValid || !comparison.RightValid {
		t.Fatalf("valid snapshots reported invalid: %#v", comparison)
	}
	if !comparison.Equivalent() || !comparison.SemanticEqual || !comparison.ProvenanceEqual {
		t.Fatalf("equivalent snapshots did not compare equal: %#v", comparison)
	}
	if comparison.ExactEvidenceEqual {
		t.Fatal("comparison erased the producing-host distinction")
	}
}

func TestCompareIRDoesNotPromoteInvalidSnapshots(t *testing.T) {
	valid := selfHostingIR(t, GoHostedCompilerID)
	invalid := IR{}
	comparison := CompareIR(valid, invalid)
	if !comparison.LeftValid || comparison.RightValid {
		t.Fatalf("validity was not reported independently: %#v", comparison)
	}
	if comparison.Equivalent() || comparison.SemanticEqual || comparison.ProvenanceEqual {
		t.Fatalf("invalid snapshot was treated as equivalent: %#v", comparison)
	}
	if comparison.RightError == "" {
		t.Fatal("invalid snapshot did not retain a diagnostic")
	}
}
