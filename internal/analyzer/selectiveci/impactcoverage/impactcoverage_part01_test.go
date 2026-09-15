package impactcoverage

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"testing"
)

type wantVector struct {
	decision Decision
	reason   Reason
	full     bool
	changed  uint64
	covered  uint64
	open     uint64
	bindings uint64
	work     uint64
	ids      []string
	paths    []string
}

func TestLiteralFixtureMatrix(t *testing.T) {
	cases := literalCases(t)
	for _, fixture := range cases {
		t.Run(fixture.name, func(t *testing.T) {
			got := Observe(NewInput(&fixture.base, &fixture.head))
			assertVector(t, got, fixture.want)
			assertDigests(t, fixture.base, fixture.head, got)
			t.Logf("vector changed=%d covered=%d uncovered=%d bindings=%d work=%d ids=%v paths=%v",
				got.ChangedBlobCount, got.CoveredChangedBlobCount, got.UncoveredChangedBlobCount,
				got.ChangedBindingCount, got.DeterministicWorkUnits, got.ChangedStableIDs, got.UncoveredPaths)
			t.Logf("digests base=%s head=%s source-map=%s/%s registry=%s/%s input=%s output=%s",
				got.BaseSnapshotDigest, got.HeadSnapshotDigest, got.BaseSourceMapDigest,
				got.HeadSourceMapDigest, got.BaseRegistryDigest, got.HeadRegistryDigest,
				got.InputDigest, got.OutputDigest)
		})
	}
}

type fixtureCase struct {
	name       string
	base, head selectiveci.Snapshot
	want       wantVector
}
