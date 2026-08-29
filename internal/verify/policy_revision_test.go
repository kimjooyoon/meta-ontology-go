package verify

import "testing"

func TestPolicyRevisionKeepsExistingBaselineFinding(t *testing.T) {
	current := Violation{Path: "old.go", Rule: "DAMP file lines", Actual: 90, Limit: 75}
	previous := []Violation{{Path: "old.go", Rule: "DAMP file lines", Actual: 90, Limit: 75}}
	if policyViolationRegressed([]Violation{current}, previous) {
		t.Fatal("unchanged baseline finding became a regression")
	}
}

func TestPolicyRevisionAllowsReducedExistingFinding(t *testing.T) {
	current := Violation{Path: "old.go", Rule: "DAMP file lines", Actual: 80, Limit: 75}
	previous := []Violation{{Path: "old.go", Rule: "DAMP file lines", Actual: 90, Limit: 75}}
	if policyViolationRegressed([]Violation{current}, previous) {
		t.Fatal("reduced baseline finding became a regression")
	}
}

func TestPolicyRevisionRejectsChangedFinding(t *testing.T) {
	current := Violation{Path: "old.go", Rule: "DAMP file lines", Actual: 91, Limit: 75}
	previous := []Violation{{Path: "old.go", Rule: "DAMP file lines", Actual: 90, Limit: 75}}
	if !policyViolationRegressed([]Violation{current}, previous) {
		t.Fatal("increased baseline finding was accepted")
	}
}

func TestPolicyRevisionRejectsNewFinding(t *testing.T) {
	current := Violation{Path: "new.go", Rule: "DAMP file lines", Actual: 76, Limit: 75}
	if !policyViolationRegressed([]Violation{current}, nil) {
		t.Fatal("new cap finding was treated as baseline")
	}
}

func TestPolicyRevisionComparesDuplicateFunctionLiteralsAsAMultiset(t *testing.T) {
	current := []Violation{
		{Path: "same.go", Rule: "DRY function lines", Actual: 80, Limit: 75, Detail: "function literal"},
		{Path: "same.go", Rule: "DRY function lines", Actual: 90, Limit: 75, Detail: "function literal"},
	}
	previous := []Violation{
		{Path: "same.go", Rule: "DRY function lines", Actual: 80, Limit: 75, Detail: "function literal"},
		{Path: "same.go", Rule: "DRY function lines", Actual: 90, Limit: 75, Detail: "function literal"},
	}
	if policyViolationRegressed(current, previous) {
		t.Fatal("unchanged duplicate function literals became a regression")
	}
}

func TestPolicyRevisionRejectsDuplicateKeyCountIncrease(t *testing.T) {
	current := []Violation{
		{Path: "same.go", Rule: "DRY function lines", Actual: 80, Limit: 75, Detail: "method Run"},
		{Path: "same.go", Rule: "DRY function lines", Actual: 81, Limit: 75, Detail: "method Run"},
	}
	previous := []Violation{{Path: "same.go", Rule: "DRY function lines", Actual: 80, Limit: 75, Detail: "method Run"}}
	if !policyViolationRegressed(current, previous) {
		t.Fatal("duplicate method key count increase was accepted")
	}
}

func TestPolicyRevisionRejectsSameKeyRankIncrease(t *testing.T) {
	current := []Violation{
		{Path: "same.go", Rule: "DRY function lines", Actual: 85, Limit: 75, Detail: "function literal"},
		{Path: "same.go", Rule: "DRY function lines", Actual: 90, Limit: 75, Detail: "function literal"},
	}
	previous := []Violation{
		{Path: "same.go", Rule: "DRY function lines", Actual: 80, Limit: 75, Detail: "function literal"},
		{Path: "same.go", Rule: "DRY function lines", Actual: 95, Limit: 75, Detail: "function literal"},
	}
	if !policyViolationRegressed(current, previous) {
		t.Fatal("one same-key finding increase was hidden by another decrease")
	}
}
