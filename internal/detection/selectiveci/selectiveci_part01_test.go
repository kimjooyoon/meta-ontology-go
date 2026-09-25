package selectiveci

import (
	"reflect"
	"testing"
)

func TestPlanContractTable(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		status Status
		reason string
	}{
		{name: "exact complete input", status: StatusSelective},
		{name: "unsupported schema", mutate: func(input *Input) { input.SchemaVersion = "gooo/selective-ci/v0" }, status: StatusFullSuiteFallback, reason: ReasonUnsupportedSchema},
		{name: "mismatched head digest", mutate: func(input *Input) { input.Head.Digest = digest("wrong") }, status: StatusFullSuiteFallback, reason: ReasonMismatchedDigest},
		{name: "invalid argv", mutate: func(input *Input) { input.Registry.Commands[0].Argv = nil }, status: StatusFullSuiteFallback, reason: ReasonInvalidArgv},
		{name: "resource arithmetic", mutate: func(input *Input) { input.Receipts[0].Envelope.Samples[3].WallNS = 0 }, status: StatusFullSuiteFallback, reason: ReasonResourceArithmetic},
		{name: "missing provenance", mutate: func(input *Input) { input.ProvenancePaths = input.ProvenancePaths[:1] }, status: StatusFullSuiteFallback, reason: ReasonAmbiguousPath},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := completeInput()
			if test.mutate != nil {
				test.mutate(&input)
				if test.name == "invalid argv" {
					input.Registry.Digest = input.Registry.ComputedDigest()
				}
			}
			got := Plan(input)
			if got.Status != test.status || got.ReasonCode != test.reason {
				t.Fatalf("status/reason = %s/%s, want %s/%s", got.Status, got.ReasonCode, test.status, test.reason)
			}
			if got.CanonicalDigest == "" || got.CanonicalDigest != got.StableDigest() || got.Digest != got.CanonicalDigest {
				t.Fatalf("result is not sealed: %#v", got)
			}
			if got.Status == StatusFullSuiteFallback && len(got.SelectedCommandIDs)+len(got.SelectedGuardCommandIDs) != 0 {
				t.Fatalf("fallback retained partial selection: %#v", got)
			}
		})
	}
}
func TestManifestDeletionUsesStableIDUnion(t *testing.T) {
	input := completeInput()
	input.Head.Files = []SnapshotFile{}
	input.Head.Digest = input.Head.ComputedDigest()
	for i := range input.Receipts {
		input.Receipts[i].SnapshotDigest = input.Head.Digest
	}
	got := Plan(input)
	if got.Status != StatusSelective {
		t.Fatalf("deletion status = %s/%s", got.Status, got.ReasonCode)
	}
	if !reflect.DeepEqual(got.ChangedSemanticIDs, []string{"urn:selectiveci:entity/order"}) {
		t.Fatalf("changed IDs = %#v", got.ChangedSemanticIDs)
	}
}
