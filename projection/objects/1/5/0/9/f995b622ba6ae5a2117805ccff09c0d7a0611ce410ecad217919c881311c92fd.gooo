package r4safe_workid

import (
	"testing"
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
