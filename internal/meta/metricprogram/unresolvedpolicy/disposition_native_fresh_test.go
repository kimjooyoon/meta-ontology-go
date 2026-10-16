package unresolvedpolicy

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNativeDispositionRecordFresh1864(t *testing.T) {
	for _, row := range []struct {
		name                 string
		registered, recorded bool
		want                 SourceDisposition
	}{
		{"registered-origin", true, true, SourceDisposition{
			Resolution: "exact_operation", Reason: "REGISTERED_SOURCE_OPERATION", NextOperation: "CONTINUE_SOURCE_OPERATION"}},
		{"unregistered-origin", false, true, SourceDisposition{
			Resolution: "invariant_only", Reason: "OPERATION_NOT_IN_SOURCE_REGISTRY", NextOperation: "DECLARE_SOURCE_OPERATION"}},
		{"registered-missing-origin", true, false, SourceDisposition{
			Resolution: "invariant_only", Reason: "OPERATION_ORIGIN_REQUIRED", NextOperation: "CAPTURE_OPERATION_ORIGIN"}},
		{"unregistered-missing-origin", false, false, SourceDisposition{
			Resolution: "invariant_only", Reason: "OPERATION_ORIGIN_REQUIRED", NextOperation: "CAPTURE_OPERATION_ORIGIN"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := sourceOperationDisposition(row.registered, row.recorded); got != row.want {
				t.Fatalf("native six-field source record: %+v, want %+v", got, row.want)
			}
		})
	}
}

func TestDispositionOriginBindingFresh1865(t *testing.T) {
	if err := VerifySource(os.DirFS("../../../..")); err != nil {
		t.Fatal(err)
	}
	refusal, err := NewOperationRefusal(OperationOrigin{RequestedOperation: "urn:native1865:unregistered"})
	if err != nil {
		t.Fatal(err)
	}
	want := sourceOperationDisposition(false, false)
	if refusal.Resolution != want.Resolution || refusal.Reason != want.Reason ||
		refusal.NextOperation != want.NextOperation || refusal.ReadinessCredit != want.ReadinessCredit ||
		refusal.RepositoryWorkspaceWrites != want.RepositoryWorkspaceWrites ||
		refusal.PromotionAuthorized != want.PromotionAuthorized || refusal.Policy != Origin() {
		t.Fatal("refusal did not transport the complete source record and compiled origin")
	}
	raw, err := json.Marshal(refusal)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"readiness_credit", "repository_workspace_writes", "promotion_authorized"} {
		if string(fields[name]) != "false" {
			t.Fatalf("source-owned diagnostic value missing in wire record: %s", name)
		}
	}
	t.Log("NATIVE_DISPOSITION_ORIGIN source record fields=6/6; native Boolean inputs=4; field values=24; source inputs overlap declared cases; not held-out coverage")
}
