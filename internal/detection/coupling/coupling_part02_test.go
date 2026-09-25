package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestZeroChangeManifestYieldsEmptyPass(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	entry := fixture.input.Manifest.Entries[0]
	entry.BeforeBlobDigest, entry.AfterBlobDigest = fixtureDigest("same-blob"), fixtureDigest("same-blob")
	fixture.input.Manifest.Entries = []ManifestEntry{entry}
	fixture.input.Manifest.ZeroChange = true
	fixture.input.Manifest.Digest = stableDigest(manifestCanonical(fixture.input.Manifest))
	fixture.input.Receipts = nil
	fixture.input.InferencePath = semantic.InferencePathV1{}
	result := Evaluate(fixture.input, fixture.authorityContext)
	if result.Status != StatusPass || len(result.AcceptedSurfaceIDs) != 0 {
		t.Fatalf("zero-change result = %#v", result)
	}
}
func TestProducerCannotSelfBindEmptyApplicabilityOrResourceWaiver(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	missingAuthority := Evaluate(fixture.input, AuthorityContext{})
	if missingAuthority.Status != StatusUnknown || missingAuthority.Reasons[0].Code != ReasonAuthorityInputSelfBound {
		t.Fatalf("missing evaluator authority = %#v", missingAuthority)
	}
	input := fixture.input
	input.Registry.Surfaces = []Surface{}
	input.Registry.Digest = stableDigest(registryCanonical(input.Registry))
	input.Config.RegistryDigest = input.Registry.Digest
	input.Manifest.RegistryDigest = input.Registry.Digest
	input.Manifest.Entries = []ManifestEntry{}
	input.Manifest.ZeroChange = true
	input.Manifest.Digest = stableDigest(manifestCanonical(input.Manifest))
	input.Receipts = nil
	input.InferencePath = semantic.InferencePathV1{}
	input.ExternalReceipt = nil
	got := Evaluate(input, fixture.authorityContext)
	if got.Status != StatusUnknown || got.Reasons[0].Code != ReasonRequiredInputMissing {
		t.Fatalf("empty self-bound packet = %#v", got)
	}

	waived := newFixture(t, ChangeClaimDelta)
	waived.input.Config.ExternalReceiptRequired = false
	waived.input.ExternalReceipt = nil
	got = Evaluate(waived.input, waived.authorityContext)
	if got.Status == StatusPass || got.Reasons[0].Code != ReasonAuthorityInputSelfBound {
		t.Fatalf("resource waiver packet = %#v", got)
	}
}
func TestProducerCannotRebindExternalAuthorityByRehashing(t *testing.T) {
	fixture := newFixture(t, ChangeClaimDelta)
	fixture.input.Config.ExpectedProviderDigest = fixtureDigest("attacker-provider")
	fixture.input.ExternalReceipt.ProviderDigest = fixtureDigest("attacker-provider")
	fixture.input.ExternalReceipt.Digest = stableDigest(externalCanonical(*fixture.input.ExternalReceipt))
	got := Evaluate(fixture.input, fixture.authorityContext)
	if got.Status != StatusFailClosed || got.Reasons[0].Code != ReasonAuthorityInputSelfBound {
		t.Fatalf("rehashed external authority packet = %#v", got)
	}
}
