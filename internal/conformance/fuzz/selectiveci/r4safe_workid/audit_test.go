package r4safe_workid

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

const (
	derivedWorkID = "e32d42da044475b5c74a765ebf8e6725f99674c06db1d65f3dcdf41baaa91c1d"
	zeroWorkID    = "0000000000000000000000000000000000000000000000000000000000000000"
	collisionID   = "36bbe50ed96841d10443bcb670d6554f0a34b761be67ec9c4a8ad2c0c44ca42c"
)

type auditCase struct {
	name          string
	input         Input
	expected      Result
	expectedLabel string
}

func baseInput() Input {
	return Input{SnapshotDigest: "snap", ObligationID: "obl", PathID: "path", PolicyDigest: "pol"}
}

func auditCases() []auditCase {
	base := baseInput()
	return []auditCase{
		{"A-derived", base, Result{Decision: DecisionPass, Reason: ReasonDerived, LegacyWorkID: derivedWorkID, CanonicalDigest: "9f5731b7efe8e32f42a1129bb1d2fd116e362fff5fa29a92008eeed4a26727b2"}, "pass"},
		{"B-matching", Input{SnapshotDigest: "snap", ObligationID: "obl", PathID: "path", PolicyDigest: "pol", CallerWorkID: derivedWorkID}, Result{Decision: DecisionPass, Reason: ReasonMatchingCallerOverride, LegacyWorkID: derivedWorkID, CanonicalDigest: "a6ab98ba47a8056b37326f5b304477ea671cae505199654e659f477ac1b7bb86"}, "match"},
		{"C-malformed", Input{SnapshotDigest: "snap", ObligationID: "obl", PathID: "path", PolicyDigest: "pol", CallerWorkID: "not-a-work-id"}, Result{Decision: DecisionFailClosed, Reason: ReasonMalformedOverride, CanonicalDigest: "ad06a07e4aded1307a605365da3f89b909a529f9b2bda5bf1805704a262aa065"}, "malformed"},
		{"D-forged", Input{SnapshotDigest: "snap", ObligationID: "obl", PathID: "path", PolicyDigest: "pol", CallerWorkID: zeroWorkID}, Result{Decision: DecisionFailClosed, Reason: ReasonWorkIDMismatch, CanonicalDigest: "c3cfb59e1e537b17394846643c42d5970b280973bed74a0d4522c707b54220c5"}, "forged"},
		{"E-missing-snapshot", Input{ObligationID: "obl", PathID: "path", PolicyDigest: "pol"}, Result{Decision: DecisionUnknown, Reason: ReasonRequiredInputMissing, FullSuiteRequired: true, CanonicalDigest: "ee0db8b7edbe04b3406ab209c3b4c04e4670df468441469d70956ca31a0e9206"}, "missing-snapshot"},
		{"F-missing-obligation", Input{SnapshotDigest: "snap", PathID: "path", PolicyDigest: "pol"}, Result{Decision: DecisionUnknown, Reason: ReasonRequiredInputMissing, FullSuiteRequired: true, CanonicalDigest: "ee0db8b7edbe04b3406ab209c3b4c04e4670df468441469d70956ca31a0e9206"}, "missing-obligation"},
		{"G-missing-path", Input{SnapshotDigest: "snap", ObligationID: "obl", PolicyDigest: "pol"}, Result{Decision: DecisionUnknown, Reason: ReasonRequiredInputMissing, FullSuiteRequired: true, CanonicalDigest: "ee0db8b7edbe04b3406ab209c3b4c04e4670df468441469d70956ca31a0e9206"}, "missing-path"},
		{"H-missing-policy", Input{SnapshotDigest: "snap", ObligationID: "obl", PathID: "path"}, Result{Decision: DecisionUnknown, Reason: ReasonRequiredInputMissing, FullSuiteRequired: true, CanonicalDigest: "ee0db8b7edbe04b3406ab209c3b4c04e4670df468441469d70956ca31a0e9206"}, "missing-policy"},
	}
}

func TestAuditVectors(t *testing.T) {
	for _, testCase := range auditCases() {
		if got := Audit(testCase.input); got != testCase.expected {
			t.Fatalf("%s: got %#v, want %#v", testCase.name, got, testCase.expected)
		}
	}
}

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
	for index := len(cases) - 1; index >= 0; index-- {
		if got := Audit(cases[index].input); got != baseline[cases[index].name] {
			t.Fatalf("permutation changed %s: got %#v, want %#v", cases[index].name, got, baseline[cases[index].name])
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

func TestProductionOverrideDisagreement(t *testing.T) {
	base := baseInput()
	if got := productionWorkIDFor(base.SnapshotDigest, base.ObligationID, base.PathID, base.PolicyDigest, ""); got != derivedWorkID {
		t.Fatalf("production empty override = %q, want %q", got, derivedWorkID)
	}
	if got := productionWorkIDFor(base.SnapshotDigest, base.ObligationID, base.PathID, base.PolicyDigest, derivedWorkID); got != derivedWorkID {
		t.Fatalf("production matching override = %q, want %q", got, derivedWorkID)
	}
	if got := productionWorkIDFor(base.SnapshotDigest, base.ObligationID, base.PathID, base.PolicyDigest, zeroWorkID); got != zeroWorkID {
		t.Fatalf("production forged override = %q, want %q", got, zeroWorkID)
	}
	got := Audit(Input{SnapshotDigest: base.SnapshotDigest, ObligationID: base.ObligationID, PathID: base.PathID, PolicyDigest: base.PolicyDigest, CallerWorkID: zeroWorkID})
	want := auditCases()[3].expected
	if got != want {
		t.Fatalf("audit forged override = %#v, want %#v", got, want)
	}
}
