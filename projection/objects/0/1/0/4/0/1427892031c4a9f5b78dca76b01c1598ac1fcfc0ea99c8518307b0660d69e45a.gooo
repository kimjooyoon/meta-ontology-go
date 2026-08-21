package pressureshadow

import (
	"reflect"
	"testing"
)

func assertPathVector(t *testing.T, got Result, decision Decision, reason Reason,
	missing, orphan, missingBinding, mismatch []string) {
	t.Helper()
	if got.Decision != decision || got.Reason != reason {
		t.Fatalf("decision/reason = %s/%s, want %s/%s", got.Decision, got.Reason, decision, reason)
	}
	if !sameIDs(got.MissingPathIDs, missing) || !sameIDs(got.OrphanPathIDs, orphan) ||
		!sameIDs(got.MissingBindingPathIDs, missingBinding) || !sameIDs(got.BindingMismatchPathIDs, mismatch) {
		t.Fatalf("path sets = %+v", got)
	}
	if got.EnforcementEffect != EnforcementNoEffect || got.InputDigest == "" ||
		got.ResultDigest == "" || got.ReplayDigest == "" {
		t.Fatalf("incomplete digest/effect result: %+v", got)
	}
}
func sameIDs(got, want []string) bool {
	return len(got) == 0 && len(want) == 0 || reflect.DeepEqual(got, want)
}
func a2aInput(t *testing.T) Input {
	t.Helper()
	input, err := DecodeInput([]byte(a2aRawInput))
	if err != nil {
		t.Fatal(err)
	}
	return input
}
