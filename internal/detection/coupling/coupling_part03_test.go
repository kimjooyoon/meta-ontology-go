package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestExplicitEmptyApplicabilityProofIsEvaluatorOwned(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	authority := fixture.authorityContext
	authority.Registry = Registry{Schema: RegistrySchemaV1, Surfaces: []Surface{}}
	authority.Registry.Digest = stableDigest(registryCanonical(authority.Registry))
	proof := ApplicabilityProof{
		Schema: AuthorityContextSchemaV1, RegistryDigest: authority.Registry.Digest,
		ToolchainDigest: authority.ToolchainDigest, ProfileDigest: authority.ProfileDigest,
		SnapshotDigest: authority.SnapshotDigest, AllowsEmpty: true,
	}
	proof.Digest = stableDigest(applicabilityCanonical(proof))
	authority.Applicability = &proof
	input := fixture.input
	input.Registry = authority.Registry
	input.Config.RegistryDigest = authority.Registry.Digest
	input.Manifest = ChangeManifest{
		Schema: ManifestSchemaV1, Complete: true, ZeroChange: true,
		RegistryDigest: authority.Registry.Digest, ToolchainDigest: authority.ToolchainDigest,
		ProfileDigest: authority.ProfileDigest, BeforeSnapshotDigest: fixtureDigest("snapshot-before"),
		AfterSnapshotDigest: authority.SnapshotDigest, Entries: []ManifestEntry{},
	}
	input.Manifest.Digest = stableDigest(manifestCanonical(input.Manifest))
	input.Receipts = nil
	input.InferencePath = semantic.InferencePathV1{}
	input.ExternalReceipt = externalReceiptFor(input.Config)
	got := Evaluate(input, authority)
	if got.Status != StatusPass || len(got.AcceptedSurfaceIDs) != 0 {
		t.Fatalf("explicit empty applicability result = %#v", got)
	}
}
func TestPresentationAndRelocationDoNotChangeCanonicalDecision(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	baseline := Evaluate(fixture.input, fixture.authorityContext)
	fixture.input.Registry.Surfaces[0].PresentationLabel = "RenamedPayOrder"
	fixture.input.Registry.Surfaces[0].Binding.PackageLabel = "renamed.billing"
	fixture.input.Registry.Surfaces[0].Binding.FileLabel = "relocated.go"
	fixture.input.WorkspaceRoot = "/another/workspace"
	fixture.input.Manifest.Entries[0].BeforeSourcePath = "/different/old.go"
	fixture.input.Manifest.Entries[0].AfterSourcePath = "/different/new.go"
	changed := Evaluate(fixture.input, fixture.authorityContext)
	if changed.Status != baseline.Status || changed.Digest != baseline.Digest || changed.InputDigest != baseline.InputDigest || !reflect.DeepEqual(changed.AcceptedSurfaceIDs, baseline.AcceptedSurfaceIDs) {
		t.Fatalf("presentation mutation changed result: before=%#v after=%#v", baseline, changed)
	}
	if stableDigest(registryCanonical(fixture.input.Registry)) != fixture.input.Registry.Digest {
		t.Fatal("presentation mutation changed registry identity digest")
	}
	delta := newFixture(t, ChangeClaimDelta)
	deltaBaseline := Evaluate(delta.input, delta.authorityContext)
	delta.input.Manifest.Entries[0].BeforeSourcePath = "/other/root/before.go"
	delta.input.Manifest.Entries[0].AfterSourcePath = "/other/root/after.go"
	delta.input.Receipts[0].AuthoritativeSource.Path = "renamed/authority.gooo"
	delta.input.Receipts[0].AuthoritativeSource.Span = "99:1-99:2"
	delta.input.WorkspaceRoot = "/other/root"
	deltaChanged := Evaluate(delta.input, delta.authorityContext)
	if deltaChanged.Status != deltaBaseline.Status || deltaChanged.Digest != deltaBaseline.Digest || deltaChanged.InputDigest != deltaBaseline.InputDigest {
		t.Fatalf("source presentation mutation changed result: before=%#v after=%#v", deltaBaseline, deltaChanged)
	}
}
