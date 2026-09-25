package selectiveci

import (
	"reflect"
	"testing"
)

func TestDiffChangedDeletedRelocatedAndUnchanged(t *testing.T) {
	base := testSnapshot(t, "pkg/old.go", "Order", "urn:gooo:entity:order")
	unchanged := testSnapshot(t, "pkg/old.go", "Order", "urn:gooo:entity:order")
	renamed := testSnapshot(t, "pkg/new.go", "RenamedOrder", "urn:gooo:entity:order")
	deleted, err := Build(SnapshotInput{
		SourceMapDigest: testDigest("source-map"),
		RegistryDigest:  testDigest("registry"),
		RegisteredIDs:   []string{"urn:gooo:entity:order"},
	})
	if err != nil {
		t.Fatalf("empty head snapshot: %v", err)
	}

	cases := []struct {
		name string
		head Snapshot
		want []string
	}{
		{name: "identical exact binding", head: unchanged, want: []string{}},
		{name: "rename", head: renamed, want: []string{"urn:gooo:entity:order"}},
		{name: "deletion", head: deleted, want: []string{"urn:gooo:entity:order"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Diff(base, tc.head)
			if err != nil {
				t.Fatalf("Diff: %v", err)
			}
			if got.Status != StatusBound || got.FullSuiteFallback || !reflect.DeepEqual(got.ChangedIDs, tc.want) {
				t.Fatalf("delta = %#v, want IDs %#v", got, tc.want)
			}
		})
	}
}
func TestDiffRelocationIncludesStableID(t *testing.T) {
	base := testSnapshot(t, "pkg/old.go", "Order", "urn:gooo:entity:order")
	relocated := testSnapshot(t, "internal/order.go", "Order", "urn:gooo:entity:order")
	got, err := Diff(base, relocated)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !reflect.DeepEqual(got.ChangedIDs, []string{"urn:gooo:entity:order"}) {
		t.Fatalf("relocation IDs = %#v", got.ChangedIDs)
	}
}
