package couplingexplain

import (
	"context"
)

type NoLink struct {
	Reason LinkReason `json:"reason"`
}

// VerifiedEnvelopeAdapter is the future detector/oracle integration seam.
// Adapters must return an envelope whose Verdict and digests are already
// authoritative; they may not be implemented by this view package.
type VerifiedEnvelopeAdapter interface {
	DecodeVerifiedEnvelope([]byte) (VerifiedEnvelope, error)
}

// CanonicalInputs names the five source documents owned by the future
// detector/oracle adapter. This package deliberately does not reconcile them.
type CanonicalInputs struct {
	Registry        []byte
	Bindings        []byte
	Paths           []byte
	Receipts        []byte
	VerifierResults []byte
}

// DetectorOracleAdapter is the integration seam for the immutable upstream
// detector/oracle envelope. Its implementation owns semantic reconciliation.
type DetectorOracleAdapter interface {
	VerifyCanonicalSnapshot(CanonicalInputs) (VerifiedEnvelope, error)
}

func ExplainEnvelopeBytes(ctx context.Context, request Request, data []byte) (Explanation, error) {
	envelope, err := DecodeVerifiedEnvelope(data)
	if err != nil {
		return Explanation{}, err
	}
	return Explain(ctx, request, envelope), nil
}
func ExplainWithAdapter(ctx context.Context, request Request, data []byte, adapter VerifiedEnvelopeAdapter) (Explanation, error) {
	if adapter == nil {
		return requestNoLink(requestBinding(request), StatusUnknown, ReasonMissing, "missing-envelope-adapter"), nil
	}
	envelope, err := adapter.DecodeVerifiedEnvelope(data)
	if err != nil {
		return Explanation{}, err
	}
	return Explain(ctx, request, envelope), nil
}
func ExplainCanonicalSnapshot(ctx context.Context, request Request, inputs CanonicalInputs, adapter DetectorOracleAdapter) (Explanation, error) {
	if adapter == nil {
		return requestNoLink(requestBinding(request), StatusUnknown, ReasonMissing, "missing-canonical-adapter"), nil
	}
	envelope, err := adapter.VerifyCanonicalSnapshot(inputs)
	if err != nil {
		return Explanation{}, err
	}
	return Explain(ctx, request, envelope), nil
}
