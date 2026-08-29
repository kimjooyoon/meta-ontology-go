package verify

import "testing"

func TestPolicyRevisionKeepsExistingBaselineFinding(t *testing.T) {
	current := Violation{Path: "old.go", Rule: "DAMP file lines", Actual: 90, Limit: 75}
	previous := []Violation{{Path: "old.go", Rule: "DAMP file lines", Actual: 90, Limit: 75}}
	if policyViolationRegressed(current, previous) {
		t.Fatal("unchanged baseline finding became a regression")
	}
}

func TestPolicyRevisionAllowsReducedExistingFinding(t *testing.T) {
	current := Violation{Path: "old.go", Rule: "DAMP file lines", Actual: 80, Limit: 75}
	previous := []Violation{{Path: "old.go", Rule: "DAMP file lines", Actual: 90, Limit: 75}}
	if policyViolationRegressed(current, previous) {
		t.Fatal("reduced baseline finding became a regression")
	}
}

func TestPolicyRevisionRejectsChangedFinding(t *testing.T) {
	current := Violation{Path: "old.go", Rule: "DAMP file lines", Actual: 91, Limit: 75}
	previous := []Violation{{Path: "old.go", Rule: "DAMP file lines", Actual: 90, Limit: 75}}
	if !policyViolationRegressed(current, previous) {
		t.Fatal("increased baseline finding was accepted")
	}
}

func TestPolicyRevisionRejectsNewFinding(t *testing.T) {
	current := Violation{Path: "new.go", Rule: "DAMP file lines", Actual: 76, Limit: 75}
	if !policyViolationRegressed(current, nil) {
		t.Fatal("new cap finding was treated as baseline")
	}
}

func TestPolicyRevisionRejectsFalseClosure(t *testing.T) {
	if err := ValidatePolicyClosure(true, false); err == nil {
		t.Fatal("closed claim without receipt was accepted")
	}
}
