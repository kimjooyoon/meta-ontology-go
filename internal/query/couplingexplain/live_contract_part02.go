package couplingexplain

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/couplingmanifest"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

// MissingLiveLinkEnvelope records the exact upstream decision and withholds
// the link because the current live surface has no term/path/oracle payload.
func MissingLiveLinkEnvelope(binding SnapshotBinding, snapshot LiveSnapshot) (VerifiedEnvelope, error) {
	if err := validateLiveSnapshot(snapshot); err != nil {
		return VerifiedEnvelope{}, err
	}
	if binding.SourceMapDigest != snapshot.ManifestMetadata.SourceMapDigest ||
		binding.ManifestDigest != snapshot.Manifest.Digest ||
		binding.DetectorInputDigest != snapshot.DetectorResult.InputDigest ||
		binding.DetectorResultDigest != snapshot.DetectorResult.Digest {
		return VerifiedEnvelope{}, fmt.Errorf("live snapshot digests are not bound into query request")
	}
	binding.EnvelopeDigest = ""
	envelope := VerifiedEnvelope{
		Schema:         "gooo-coupling-explanation/v1",
		Binding:        binding,
		Upstream:       upstreamEvidence(snapshot),
		Verdict:        VerdictUnknown,
		NoLinkReason:   ReasonMissing,
		EvidenceDigest: snapshot.DetectorResult.Digest,
		Diagnostics:    []Diagnostic{{Code: "missing-live-term-path-verifier", IDs: acceptedIDs(snapshot.DetectorResult)}},
	}
	envelope.EnvelopeDigest = envelope.Digest()
	return envelope, nil
}
func validateLiveSnapshot(snapshot LiveSnapshot) error {
	if snapshot.Manifest.Schema != detector.ManifestSchemaV1 {
		return fmt.Errorf("live manifest schema is malformed")
	}
	if !validDigest(snapshot.Manifest.Digest) {
		return fmt.Errorf("live manifest digest is malformed")
	}
	if snapshot.ManifestMetadata.SourceMapDigest != "" && !validDigest(snapshot.ManifestMetadata.SourceMapDigest) {
		return fmt.Errorf("live source-map digest is malformed")
	}
	if snapshot.DetectorInput.Manifest.Digest != snapshot.Manifest.Digest {
		return fmt.Errorf("live manifest digest is not bound into detector input")
	}
	if _, err := detectorResultBytes(snapshot.DetectorResult); err != nil {
		return err
	}
	return nil
}
func detectorResultBytes(result detector.Result) ([]byte, error) {
	data, err := couplingmanifest.EncodeResult(result)
	if err != nil {
		return nil, err
	}
	if _, err := couplingmanifest.DecodeResult(data); err != nil {
		return nil, fmt.Errorf("live detector result: %w", err)
	}
	return data, nil
}
func upstreamEvidence(snapshot LiveSnapshot) *UpstreamEvidence {
	return &UpstreamEvidence{
		SourceMapDigest:      snapshot.ManifestMetadata.SourceMapDigest,
		ManifestDigest:       snapshot.Manifest.Digest,
		DetectorInputDigest:  snapshot.DetectorResult.InputDigest,
		DetectorResultDigest: snapshot.DetectorResult.Digest,
		DetectorStatus:       snapshot.DetectorResult.Status,
		DetectorReasons:      append([]detector.Reason(nil), snapshot.DetectorResult.Reasons...),
		ManifestStatus:       snapshot.ManifestMetadata.Status,
		ManifestReason:       snapshot.ManifestMetadata.Reason,
	}
}
