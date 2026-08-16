package couplingexplain

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/couplingmanifest"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

// UpstreamEvidence preserves live detector and manifest decisions without
// translating them into query-local truth. It is part of explanation identity.
type UpstreamEvidence struct {
	SourceMapDigest      string                              `json:"source_map_digest,omitempty"`
	ManifestDigest       string                              `json:"manifest_digest"`
	DetectorInputDigest  string                              `json:"detector_input_digest"`
	DetectorResultDigest string                              `json:"detector_result_digest"`
	DetectorStatus       detector.Status                     `json:"detector_status"`
	DetectorReasons      []detector.Reason                   `json:"detector_reasons,omitempty"`
	ManifestStatus       couplingmanifest.ConstructionStatus `json:"manifest_status,omitempty"`
	ManifestReason       couplingmanifest.ConstructionCode   `json:"manifest_reason,omitempty"`
}

// LiveSnapshot contains exact upstream values. The manifest and result types
// are aliases from the live packages; this package does not redeclare them.
type LiveSnapshot struct {
	Manifest         couplingmanifest.Manifest
	DetectorInput    detector.Input
	DetectorResult   detector.Result
	ManifestMetadata couplingmanifest.Metadata
}

// LiveSnapshotAdapter supplies an already authoritative oracle envelope. It
// must return UNKNOWN/no-link when live data lacks term, path, or verifier
// material; it may not infer those fields from names, paths, or rules.
type LiveSnapshotAdapter interface {
	AdaptLiveSnapshot(LiveSnapshot) (VerifiedEnvelope, error)
}

// ExplainLiveSnapshot projects a live detector/manifest snapshot through the
// same pure envelope checks used by every other query entry point.
func ExplainLiveSnapshot(ctx context.Context, request Request, snapshot LiveSnapshot, adapter LiveSnapshotAdapter) (Explanation, error) {
	if adapter == nil {
		return Explanation{}, fmt.Errorf("live snapshot adapter is nil")
	}
	envelope, err := adapter.AdaptLiveSnapshot(snapshot)
	if err != nil {
		return Explanation{}, err
	}
	return Explain(ctx, request, envelope), nil
}

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

func acceptedIDs(result detector.Result) []string {
	ids := make([]string, 0, len(result.AcceptedSurfaceIDs))
	for _, id := range result.AcceptedSurfaceIDs {
		ids = append(ids, id.String())
	}
	return sortedStrings(ids)
}
