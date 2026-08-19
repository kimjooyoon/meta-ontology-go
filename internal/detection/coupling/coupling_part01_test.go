package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestDeltaAndNoDeltaPositiveFixtures(t *testing.T) {
	for _, claim := range []ChangeClaim{ChangeClaimDelta, ChangeClaimNoDelta} {
		t.Run(string(claim), func(t *testing.T) {
			fixture := newFixture(t, claim)
			result := Evaluate(fixture.input, fixture.authorityContext)
			if result.Status != StatusPass || !reflect.DeepEqual(result.AcceptedSurfaceIDs, []semantic.ID{fixture.surface}) {
				t.Fatalf("result = %#v", result)
			}
			if !result.Observation.CPU.Known || !result.Observation.Memory.Known || !result.Observation.ResourceWork.Known || !result.Observation.DeterministicWork.Known || result.Observation.ResourceWork.Value != 13 {
				t.Fatalf("resource or work dimensions are not known: %#v", result.Observation)
			}
		})
	}
}
func TestInputDigestBindsSnapshotRegistryReceiptAndPath(t *testing.T) {
	base := newFixture(t, ChangeClaimDelta)
	want := Evaluate(base.input, base.authorityContext)
	mutations := []func(*Input){
		func(input *Input) { input.Config.SnapshotDigest = fixtureDigest("other-snapshot") },
		func(input *Input) { input.Registry.Digest = fixtureDigest("other-registry") },
		func(input *Input) { input.Receipts[0].DeltaDigest = fixtureDigest("other-delta") },
		func(input *Input) { input.InferencePath.Edges[1].RecordID = fixtureID("other-projection") },
	}
	for index, mutate := range mutations {
		input := newFixture(t, ChangeClaimDelta).input
		mutate(&input)
		got := Evaluate(input, base.authorityContext)
		if got.InputDigest == want.InputDigest || got.Digest == want.Digest {
			t.Fatalf("mutation %d was not bound: want input/result digest change, got %#v", index, got)
		}
	}
	otherAuthority := base.authorityContext
	otherAuthority.ExpectedObserverDigest = fixtureDigest("other-observer")
	got := Evaluate(base.input, otherAuthority)
	if got.InputDigest == want.InputDigest || got.Digest == want.Digest {
		t.Fatalf("authority mutation was not bound: got %#v", got)
	}
}
func TestExternalReceiptIsIndependentlyBound(t *testing.T) {
	missing := newFixture(t, ChangeClaimDelta)
	missing.input.ExternalReceipt.DeterministicWorkUnits = nil
	result := Evaluate(missing.input, missing.authorityContext)
	if result.Status != StatusUnknown || result.Reasons[0].Code != ReasonExternalReceiptMissing || result.Observation.CPU.Known || result.Observation.Memory.Known {
		t.Fatalf("missing external work = %#v", result)
	}
	mismatch := newFixture(t, ChangeClaimDelta)
	mismatch.input.ExternalReceipt.ProviderDigest = fixtureDigest("wrong-provider")
	mismatch.input.ExternalReceipt.Digest = stableDigest(externalCanonical(*mismatch.input.ExternalReceipt))
	result = Evaluate(mismatch.input, mismatch.authorityContext)
	if result.Status != StatusFailClosed || result.Reasons[0].Code != ReasonDigestMismatch {
		t.Fatalf("provider mismatch = %#v", result)
	}
}
func TestEvaluateIsReadOnly(t *testing.T) {
	fixture := newFixture(t, ChangeClaimDelta)
	before := fixture.input
	_ = Evaluate(fixture.input, fixture.authorityContext)
	if !reflect.DeepEqual(fixture.input, before) {
		t.Fatal("Evaluate mutated its input")
	}
}
