package couplingexplain

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

// DetectorSnapshot is the byte boundary for the immutable coupling detector
// API: one canonical coupling.Input and its independently produced
// coupling.Result. The evaluator-owned coupling.AuthorityContext is
// intentionally not represented as input bytes: the detector API does not
// decode it from an input packet. A concrete adapter must bind that context
// and the independent oracle envelope before returning a VerifiedEnvelope.
type DetectorSnapshot struct {
	InputBytes  []byte
	ResultBytes []byte
}

// DetectorResult is the exact merged detector result type; no query-local
// enum or digest vocabulary is introduced.
type DetectorResult = coupling.Result

// DecodeDetectorResult delegates strict shape and digest validation to the
// merged detector package. It still does not turn a detector PASS into an
// explanation link; that requires an already verified oracle envelope.
func DecodeDetectorResult(data []byte) (DetectorResult, error) {
	return coupling.DecodeResult(data)
}

// DetectorSnapshotAdapter is the future thin adapter over the exact detector
// package. It must return an already verified envelope or a fail-closed error.
type DetectorSnapshotAdapter interface {
	AdaptDetectorSnapshot(DetectorSnapshot) (VerifiedEnvelope, error)
}

func ExplainDetectorSnapshot(ctx context.Context, request Request, snapshot DetectorSnapshot, adapter DetectorSnapshotAdapter) (Explanation, error) {
	if adapter == nil {
		return requestNoLink(requestBinding(request), StatusUnknown, ReasonMissing, "missing-detector-adapter"), nil
	}
	envelope, err := adapter.AdaptDetectorSnapshot(snapshot)
	if err != nil {
		return Explanation{}, err
	}
	return Explain(ctx, request, envelope), nil
}
