package r4safe_workid

import (
	"testing"
)

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
