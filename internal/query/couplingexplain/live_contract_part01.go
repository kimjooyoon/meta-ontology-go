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
