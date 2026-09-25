package coupling

import "testing"

func TestAuthorityInputStatePartitions(t *testing.T) {
	cases := []struct {
		name   string
		status Status
		code   ReasonCode
		mutate func(*couplingFixture)
	}{
		{"incomplete manifest", StatusUnknown, ReasonRequiredInputMissing, func(f *couplingFixture) { f.input.Manifest.Complete = false }},
		{"missing receipt", StatusUnknown, ReasonRequiredInputMissing, func(f *couplingFixture) { f.input.Receipts = nil }},
		{"stale receipt", StatusUnknown, ReasonStaleInput, func(f *couplingFixture) { f.input.Receipts[0].State = "OLD" }},
		{"source map mismatch", StatusFailClosed, ReasonSourceMapMismatch, func(f *couplingFixture) { f.input.Receipts[0].SourceMapBindingDigest = fixtureDigest("wrong-binding") }},
		{"unregistered surface", StatusFailClosed, ReasonSurfaceUnregistered, func(f *couplingFixture) {
			f.input.Manifest.Entries[0].SurfaceID = fixtureID("unregistered")
			f.input.Manifest.Digest = stableDigest(manifestCanonical(f.input.Manifest))
		}},
		{"duplicate registry surface", StatusFailClosed, ReasonDuplicateSurface, func(f *couplingFixture) {
			f.input.Registry.Surfaces = append(f.input.Registry.Surfaces, f.input.Registry.Surfaces[0])
			f.input.Registry.Digest = stableDigest(registryCanonical(f.input.Registry))
			f.input.Config.RegistryDigest = f.input.Registry.Digest
			f.input.Manifest.RegistryDigest = f.input.Registry.Digest
			f.input.Manifest.Digest = stableDigest(manifestCanonical(f.input.Manifest))
			f.authorityContext.Registry = f.input.Registry
			f.authorityContext.Registry.Surfaces = append([]Surface(nil), f.input.Registry.Surfaces...)
		}},
		{"orphan receipt", StatusFailClosed, ReasonOrphanReceipt, func(f *couplingFixture) {
			orphan := f.input.Receipts[0]
			orphan.ReceiptID = fixtureID("orphan-receipt")
			orphan.SurfaceID = fixtureID("orphan-surface")
			f.input.Receipts = append(f.input.Receipts, orphan)
		}},
		{"duplicate receipt", StatusFailClosed, ReasonDuplicateReceipt, func(f *couplingFixture) {
			duplicate := f.input.Receipts[0]
			duplicate.ReceiptID = fixtureID("second-receipt")
			f.input.Receipts = append(f.input.Receipts, duplicate)
		}},
		{"delta without source", StatusFailClosed, ReasonDeltaWithoutSource, func(f *couplingFixture) { f.input.Receipts[0].AuthoritativeSource = nil }},
		{"no delta without equality", StatusFailClosed, ReasonNoDeltaWithoutEquality, func(f *couplingFixture) {
			f.input = newFixture(t, ChangeClaimNoDelta).input
			f.input.Receipts[0].AfterCanonicalSemanticDigest = fixtureDigest("changed-semantic")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newFixture(t, ChangeClaimDelta)
			tc.mutate(&fixture)
			result := Evaluate(fixture.input, fixture.authorityContext)
			if result.Status != tc.status || len(result.AcceptedSurfaceIDs) != 0 || len(result.Reasons) == 0 || result.Reasons[0].Code != tc.code {
				t.Fatalf("state = %#v, want %s/%s", result, tc.status, tc.code)
			}
		})
	}
}
