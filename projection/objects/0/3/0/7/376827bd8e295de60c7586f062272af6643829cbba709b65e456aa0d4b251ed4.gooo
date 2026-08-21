package couplingexplain

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/couplingmanifest"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"testing"
)

func TestLiveSnapshotPreservesDetectorDecisionAndWithholdsMissingLink(t *testing.T) {
	request, _ := fixtureEnvelope(t, ClaimNoDelta, VerdictVerified)
	result := literalDetectorResult()
	manifestDigest := request.ManifestDigest
	snapshot := LiveSnapshot{
		Manifest:         couplingmanifest.Manifest{Schema: detector.ManifestSchemaV1, Digest: manifestDigest},
		DetectorInput:    detector.Input{Manifest: detector.ChangeManifest{Digest: manifestDigest}},
		DetectorResult:   result,
		ManifestMetadata: couplingmanifest.Metadata{Status: couplingmanifest.ConstructionUnknown, Reason: couplingmanifest.CodeMissingAuthority, SourceMapDigest: digest("source-map")},
	}
	request.DetectorInputDigest = result.InputDigest
	request.DetectorResultDigest = result.Digest
	binding := SnapshotBinding{SnapshotDigest: request.SnapshotDigest, RegistryDigest: request.RegistryDigest, SourceMapDigest: digest("source-map"), ManifestDigest: manifestDigest, ToolchainDigest: request.ToolchainDigest, ProfileDigest: request.ProfileDigest, DetectorInputDigest: result.InputDigest, DetectorResultDigest: result.Digest, VerifierResultDigest: request.VerifierResultDigest, Control: request.Control}
	adapter := liveFixtureAdapter{binding: binding}
	envelope, err := adapter.AdaptLiveSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	request.EnvelopeDigest = envelope.EnvelopeDigest
	got, err := ExplainLiveSnapshot(context.Background(), request, snapshot, adapter)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusUnknown || got.Link != nil || got.NoLink == nil || got.NoLink.Reason != ReasonMissing {
		t.Fatalf("live explanation = %#v", got)
	}
	if got.Upstream == nil || got.Upstream.DetectorStatus != detector.StatusUnknown || got.Upstream.DetectorResultDigest != result.Digest || got.Upstream.ManifestDigest != manifestDigest {
		t.Fatalf("upstream evidence = %#v", got.Upstream)
	}
}
func TestLiveSnapshotRejectsTamperedOrConflictingUpstreamBytes(t *testing.T) {
	request, _ := fixtureEnvelope(t, ClaimNoDelta, VerdictVerified)
	result := literalDetectorResult()
	snapshot := LiveSnapshot{
		Manifest:       couplingmanifest.Manifest{Schema: detector.ManifestSchemaV1, Digest: request.ManifestDigest},
		DetectorInput:  detector.Input{Manifest: detector.ChangeManifest{Digest: request.ManifestDigest}},
		DetectorResult: result,
	}
	tampered := snapshot
	tampered.DetectorResult.Digest = digest("tampered-result")
	if _, err := MissingLiveLinkEnvelope(SnapshotBinding{}, tampered); err == nil {
		t.Fatal("tampered detector result accepted")
	}
	conflicting := snapshot
	conflicting.DetectorInput.Manifest.Digest = digest("other-manifest")
	if _, err := MissingLiveLinkEnvelope(SnapshotBinding{}, conflicting); err == nil {
		t.Fatal("conflicting manifest accepted")
	}
}
