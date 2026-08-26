package r4safe_workid

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"testing"
)

func TestExpectedLabelsAndPermutationDoNotAffectResults(t *testing.T) {
	cases := auditCases()
	baseline := make(map[string]Result, len(cases))
	for _, testCase := range cases {
		baseline[testCase.name] = Audit(testCase.input)
	}
	mutations := []struct {
		name   string
		mutate func(*auditCase)
	}{
		{"decision", func(testCase *auditCase) { testCase.expected.Decision = DecisionUnknown }},
		{"reason", func(testCase *auditCase) { testCase.expected.Reason = ReasonRequiredInputMissing }},
		{"work ID", func(testCase *auditCase) { testCase.expected.LegacyWorkID = LegacyWorkID("mutated") }},
		{"full suite", func(testCase *auditCase) { testCase.expected.FullSuiteRequired = true }},
		{"authorization", func(testCase *auditCase) { testCase.expected.ExecutionAuthorized = true }},
		{"effect", func(testCase *auditCase) { testCase.expected.EnforcementEffect = EnforcementEffect(255) }},
		{"digest", func(testCase *auditCase) { testCase.expected.CanonicalDigest = "mutated" }},
		{"label", func(testCase *auditCase) { testCase.expectedLabel = "mutated" }},
	}
	want := cases[0].expected
	for _, mutation := range mutations {
		mutated := cases[0]
		mutation.mutate(&mutated)
		if got := Audit(mutated.input); got != want {
			t.Fatalf("expected %s mutation changed %s: got %#v, want %#v", mutation.name, mutated.name, got, want)
		}
	}
	for _, testCase := range slices.Backward(cases) {
		if got := Audit(testCase.input); got != baseline[testCase.name] {
			t.Fatalf("permutation changed %s: got %#v, want %#v", testCase.name, got, baseline[testCase.name])
		}
	}
}
func TestGovernedInputMutationChangesResult(t *testing.T) {
	base := baseInput()
	want := Audit(base)
	mutations := []func(*Input){
		func(input *Input) { input.SnapshotDigest += "x" },
		func(input *Input) { input.ObligationID += "x" },
		func(input *Input) { input.PathID += "x" },
		func(input *Input) { input.PolicyDigest += "x" },
	}
	for _, mutate := range mutations {
		input := base
		mutate(&input)
		got := Audit(input)
		if got.Decision != DecisionPass || got.CanonicalDigest == want.CanonicalDigest || got.LegacyWorkID == want.LegacyWorkID {
			t.Fatalf("governed mutation was not observed: got %#v, want unlike %#v", got, want)
		}
	}
}
func TestUndelimitedCollisionEvidence(t *testing.T) {
	left := Audit(Input{SnapshotDigest: "ab", ObligationID: "c", PathID: "d", PolicyDigest: "e"})
	right := Audit(Input{SnapshotDigest: "a", ObligationID: "bc", PathID: "d", PolicyDigest: "e"})
	want := Result{Decision: DecisionPass, Reason: ReasonDerived, LegacyWorkID: collisionID, CanonicalDigest: "bed83c55a2352fecd1633bc64da9bfaaed44a3370cdaff52265146696eb7de00"}
	if left != want || right != want {
		t.Fatalf("legacy collision changed: left=%#v right=%#v want=%#v", left, right, want)
	}
}
func productionWorkIDFor(snapshot, obligation, path, policy, override string) string {
	input := workfrontier.Input{SnapshotDigest: snapshot, PolicyDigest: policy}
	repairPath := workfrontier.RepairPath{StableID: path, WorkID: override, ObligationID: obligation}
	return workfrontier.WorkIDFor(input, repairPath)
}
