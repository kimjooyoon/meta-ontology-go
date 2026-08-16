package coupling

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestDeltaAndNoDeltaPositiveFixtures(t *testing.T) {
	for _, claim := range []ChangeClaim{ChangeClaimDelta, ChangeClaimNoDelta} {
		t.Run(string(claim), func(t *testing.T) {
			fixture := newFixture(t, claim)
			result := Evaluate(fixture.input)
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
	want := Evaluate(base.input)
	mutations := []func(*Input){
		func(input *Input) { input.Config.SnapshotDigest = fixtureDigest("other-snapshot") },
		func(input *Input) { input.Registry.Digest = fixtureDigest("other-registry") },
		func(input *Input) { input.Receipts[0].DeltaDigest = fixtureDigest("other-delta") },
		func(input *Input) { input.InferencePath.Edges[1].RecordID = fixtureID("other-projection") },
	}
	for index, mutate := range mutations {
		input := newFixture(t, ChangeClaimDelta).input
		mutate(&input)
		got := Evaluate(input)
		if got.InputDigest == want.InputDigest || got.Digest == want.Digest {
			t.Fatalf("mutation %d was not bound: want input/result digest change, got %#v", index, got)
		}
	}
}

func TestExternalReceiptIsIndependentlyBound(t *testing.T) {
	missing := newFixture(t, ChangeClaimDelta)
	missing.input.ExternalReceipt.DeterministicWorkUnits = nil
	result := Evaluate(missing.input)
	if result.Status != StatusUnknown || result.Reasons[0].Code != ReasonExternalReceiptMissing || result.Observation.CPU.Known || result.Observation.Memory.Known {
		t.Fatalf("missing external work = %#v", result)
	}
	mismatch := newFixture(t, ChangeClaimDelta)
	mismatch.input.Config.ExpectedProviderDigest = fixtureDigest("wrong-provider")
	result = Evaluate(mismatch.input)
	if result.Status != StatusFailClosed || result.Reasons[0].Code != ReasonDigestMismatch {
		t.Fatalf("provider mismatch = %#v", result)
	}
}

func TestEvaluateIsReadOnly(t *testing.T) {
	fixture := newFixture(t, ChangeClaimDelta)
	before := fixture.input
	_ = Evaluate(fixture.input)
	if !reflect.DeepEqual(fixture.input, before) {
		t.Fatal("Evaluate mutated its input")
	}
}

func TestZeroChangeManifestYieldsEmptyPass(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	entry := fixture.input.Manifest.Entries[0]
	entry.BeforeBlobDigest, entry.AfterBlobDigest = fixtureDigest("same-blob"), fixtureDigest("same-blob")
	fixture.input.Manifest.Entries = []ManifestEntry{entry}
	fixture.input.Manifest.ZeroChange = true
	fixture.input.Manifest.Digest = stableDigest(manifestCanonical(fixture.input.Manifest))
	fixture.input.Receipts = nil
	fixture.input.InferencePath = semantic.InferencePathV1{}
	result := Evaluate(fixture.input)
	if result.Status != StatusPass || len(result.AcceptedSurfaceIDs) != 0 {
		t.Fatalf("zero-change result = %#v", result)
	}
}

func TestPresentationAndRelocationDoNotChangeCanonicalDecision(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	baseline := Evaluate(fixture.input)
	fixture.input.Registry.Surfaces[0].PresentationLabel = "RenamedPayOrder"
	fixture.input.Registry.Surfaces[0].Binding.PackageLabel = "renamed.billing"
	fixture.input.Registry.Surfaces[0].Binding.FileLabel = "relocated.go"
	fixture.input.WorkspaceRoot = "/another/workspace"
	fixture.input.Manifest.Entries[0].BeforeSourcePath = "/different/old.go"
	fixture.input.Manifest.Entries[0].AfterSourcePath = "/different/new.go"
	changed := Evaluate(fixture.input)
	if changed.Status != baseline.Status || changed.Digest != baseline.Digest || changed.InputDigest != baseline.InputDigest || !reflect.DeepEqual(changed.AcceptedSurfaceIDs, baseline.AcceptedSurfaceIDs) {
		t.Fatalf("presentation mutation changed result: before=%#v after=%#v", baseline, changed)
	}
	if stableDigest(registryCanonical(fixture.input.Registry)) != fixture.input.Registry.Digest {
		t.Fatal("presentation mutation changed registry identity digest")
	}
	delta := newFixture(t, ChangeClaimDelta)
	deltaBaseline := Evaluate(delta.input)
	delta.input.Manifest.Entries[0].BeforeSourcePath = "/other/root/before.go"
	delta.input.Manifest.Entries[0].AfterSourcePath = "/other/root/after.go"
	delta.input.Receipts[0].AuthoritativeSource.Path = "renamed/authority.gooo"
	delta.input.Receipts[0].AuthoritativeSource.Span = "99:1-99:2"
	delta.input.WorkspaceRoot = "/other/root"
	deltaChanged := Evaluate(delta.input)
	if deltaChanged.Status != deltaBaseline.Status || deltaChanged.Digest != deltaBaseline.Digest || deltaChanged.InputDigest != deltaBaseline.InputDigest {
		t.Fatalf("source presentation mutation changed result: before=%#v after=%#v", deltaBaseline, deltaChanged)
	}
}

func TestLocalizationBenefitIsFiniteAndComponentwise(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	second := fixture.input.Registry.Surfaces[0]
	second.SurfaceID, second.CodeSymbolID, second.SemanticOwnerID = fixtureID("surface-two"), fixtureID("code-two"), fixtureID("owner-two")
	second.Binding.SourceMapID = fixtureID("source-map-two")
	second.Binding.BindingDigest = bindingDigest(second)
	fixture.input.Registry.Surfaces = append(fixture.input.Registry.Surfaces, second)
	fixture.input.Registry.Digest = stableDigest(registryCanonical(fixture.input.Registry))
	fixture.input.Config.RegistryDigest = fixture.input.Registry.Digest
	fixture.input.Manifest.RegistryDigest = fixture.input.Registry.Digest
	fixture.input.Manifest.Entries = append(fixture.input.Manifest.Entries, ManifestEntry{
		SurfaceID: second.SurfaceID, CodeSymbolID: second.CodeSymbolID, SemanticOwnerID: second.SemanticOwnerID,
		BeforeBindingDigest: second.Binding.BindingDigest, AfterBindingDigest: second.Binding.BindingDigest,
		BeforeBlobDigest: fixtureDigest("second-blob"), AfterBlobDigest: fixtureDigest("second-blob"),
		BeforeSourcePath: "second-before.go", AfterSourcePath: "second-after.go",
	})
	fixture.input.Manifest.ZeroChange = false
	fixture.input.Manifest.Digest = stableDigest(manifestCanonical(fixture.input.Manifest))
	result := Evaluate(fixture.input)
	if result.Status != StatusUnknown || result.Observation.ChangedSurfaces.Value != 1 || result.Observation.Receipts.Value != 1 {
		t.Fatalf("localization baseline = %#v", result)
	}
	if result.Observation.ChangedSurfaces.Value >= uint64(len(fixture.input.Registry.Surfaces)) {
		t.Fatal("changed-surface localization was not finite")
	}
}
