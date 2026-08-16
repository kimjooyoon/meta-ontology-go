// Package couplingexplain exposes a pure, read-only view over an already
// verified detector/oracle envelope. It does not infer or authorize meaning.
package couplingexplain

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type Status string

const (
	StatusPass       Status = "PASS"
	StatusFailClosed Status = "FAIL_CLOSED"
	StatusUnknown    Status = "UNKNOWN"
)

// LinkReason is the closed set of reasons for withholding an authority link.
type LinkReason string

const (
	ReasonAmbiguous    LinkReason = "AMBIGUOUS"
	ReasonStale        LinkReason = "STALE"
	ReasonUnregistered LinkReason = "UNREGISTERED"
	ReasonMissing      LinkReason = "MISSING"
	ReasonNotVerified  LinkReason = "NOT_VERIFIED"
)

type View string

const (
	ViewCompact  View = "COMPACT"
	ViewExpanded View = "EXPANDED"
)

// ChangeClaim is deliberately separate from semantic.SemanticChangeKind.
type ChangeClaim string

const (
	ClaimDelta   ChangeClaim = "DELTA"
	ClaimNoDelta ChangeClaim = "NO_DELTA"
)

type VerifierState string

const (
	VerifierPass       VerifierState = "PASS"
	VerifierFailClosed VerifierState = "FAIL_CLOSED"
	VerifierUnknown    VerifierState = "UNKNOWN"
)

type EnvelopeVerdict string

const (
	VerdictVerified   EnvelopeVerdict = "VERIFIED"
	VerdictFailClosed EnvelopeVerdict = "FAIL_CLOSED"
	VerdictUnknown    EnvelopeVerdict = "UNKNOWN"
)

// Request binds a query to one exact verified envelope and snapshot.
type Request struct {
	CodeSymbolID         string  `json:"code_symbol_id"`
	SnapshotDigest       string  `json:"snapshot_digest"`
	RegistryDigest       string  `json:"registry_digest"`
	SourceMapDigest      string  `json:"source_map_digest"`
	ManifestDigest       string  `json:"manifest_digest"`
	ToolchainDigest      string  `json:"toolchain_digest"`
	ProfileDigest        string  `json:"profile_digest"`
	DetectorInputDigest  string  `json:"detector_input_digest"`
	DetectorResultDigest string  `json:"detector_result_digest"`
	VerifierResultDigest string  `json:"verifier_result_digest"`
	EnvelopeDigest       string  `json:"envelope_digest"`
	Control              Control `json:"control"`
}

// Control carries both requested and observed values. Any race is UNKNOWN.
type Control struct {
	RequestVersion              uint64 `json:"request_version"`
	ObservedVersion             uint64 `json:"observed_version"`
	RequestCancellationVersion  uint64 `json:"request_cancellation_version"`
	ObservedCancellationVersion uint64 `json:"observed_cancellation_version"`
}

// SnapshotBinding is stable input identity; it contains no presentation data.
type SnapshotBinding struct {
	SnapshotDigest       string  `json:"snapshot_digest"`
	RegistryDigest       string  `json:"registry_digest"`
	SourceMapDigest      string  `json:"source_map_digest"`
	ManifestDigest       string  `json:"manifest_digest"`
	ToolchainDigest      string  `json:"toolchain_digest"`
	ProfileDigest        string  `json:"profile_digest"`
	DetectorInputDigest  string  `json:"detector_input_digest"`
	DetectorResultDigest string  `json:"detector_result_digest"`
	VerifierResultDigest string  `json:"verifier_result_digest"`
	EnvelopeDigest       string  `json:"envelope_digest"`
	Control              Control `json:"control"`
}

// Presentation is accepted at the adapter boundary but never enters an
// evidence digest or canonical explanation output.
type Presentation struct {
	Label     string `json:"label,omitempty"`
	Root      string `json:"root,omitempty"`
	Path      string `json:"path,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Actor     string `json:"actor,omitempty"`
}

type CodeBindingSummary struct {
	CodeSymbolID        string       `json:"code_symbol_id"`
	SemanticOwnerID     string       `json:"semantic_owner_id"`
	RegisteredSurfaceID string       `json:"registered_surface_id"`
	SourceMapID         string       `json:"source_map_id"`
	BindingDigest       string       `json:"binding_digest"`
	CodeBindingDigest   string       `json:"code_binding_digest"`
	Presentation        Presentation `json:"presentation,omitempty"`
}

type TermSummary struct {
	TermID           string       `json:"term_id"`
	SemanticOwnerID  string       `json:"semantic_owner_id"`
	Version          string       `json:"version"`
	DefinitionDigest string       `json:"definition_digest"`
	Presentation     Presentation `json:"presentation,omitempty"`
}

type PathStep struct {
	FromID       string                  `json:"from_id"`
	ToID         string                  `json:"to_id"`
	Kind         semantic.InferenceKind  `json:"kind"`
	Phase        semantic.PhasePlacement `json:"phase"`
	RuleRef      string                  `json:"rule_ref,omitempty"`
	InputDigest  string                  `json:"input_digest"`
	OutputDigest string                  `json:"output_digest"`
	EvidenceRef  string                  `json:"evidence_ref,omitempty"`
}

type PathSummary struct {
	PathID       string       `json:"path_id"`
	StartID      string       `json:"start_id"`
	EndID        string       `json:"end_id"`
	StepCount    int          `json:"step_count"`
	PathDigest   string       `json:"path_digest"`
	Steps        []PathStep   `json:"steps,omitempty"`
	Presentation Presentation `json:"presentation,omitempty"`
}

type ReceiptSummary struct {
	ReceiptID      string                      `json:"receipt_id"`
	SurfaceID      string                      `json:"surface_id"`
	ChangeClaim    ChangeClaim                 `json:"change_claim"`
	ReceiptKind    semantic.SemanticChangeKind `json:"receipt_kind"`
	BeforeIRDigest string                      `json:"before_ir_digest"`
	AfterIRDigest  string                      `json:"after_ir_digest"`
	CanonicalDelta string                      `json:"canonical_delta,omitempty"`
	DeltaDigest    string                      `json:"delta_digest,omitempty"`
	ReceiptDigest  string                      `json:"receipt_digest"`
	OriginPathID   string                      `json:"origin_path_id"`
	EvidenceRefs   []string                    `json:"evidence_refs"`
	Presentation   Presentation                `json:"presentation,omitempty"`
}

type VerifierSummary struct {
	EvidenceID     string        `json:"evidence_id"`
	ReceiptID      string        `json:"receipt_id"`
	State          VerifierState `json:"state"`
	Independent    bool          `json:"independent"`
	EvidenceDigest string        `json:"evidence_digest"`
	VerifierDigest string        `json:"verifier_digest"`
	EvidenceRefs   []string      `json:"evidence_refs"`
	Presentation   Presentation  `json:"presentation,omitempty"`
}

// VerifiedEnvelope is the adapter seam. Its Verdict and exact digests are
// produced by the detector/oracle; this package only validates envelope
// integrity and projects an allowed link.
type VerifiedEnvelope struct {
	Schema         string             `json:"schema"`
	Binding        SnapshotBinding    `json:"binding"`
	Upstream       *UpstreamEvidence  `json:"upstream,omitempty"`
	CodeBinding    CodeBindingSummary `json:"code_binding"`
	SemanticOwner  string             `json:"semantic_owner"`
	Term           TermSummary        `json:"term"`
	OriginPath     PathSummary        `json:"origin_path"`
	Receipt        ReceiptSummary     `json:"receipt"`
	Verifier       VerifierSummary    `json:"verifier"`
	Verdict        EnvelopeVerdict    `json:"verdict"`
	NoLinkReason   LinkReason         `json:"no_link_reason,omitempty"`
	EvidenceDigest string             `json:"evidence_digest"`
	EnvelopeDigest string             `json:"envelope_digest"`
	Diagnostics    []Diagnostic       `json:"diagnostics,omitempty"`
}

// Diagnostic is intentionally stable-ID oriented; it has no free-form
// narrative where labels or paths could accidentally become authority data.
type Diagnostic struct {
	Code string   `json:"code"`
	IDs  []string `json:"ids,omitempty"`
}

type Explanation struct {
	Status         Status            `json:"status"`
	EvidenceDigest string            `json:"evidence_digest"`
	Binding        SnapshotBinding   `json:"binding"`
	Upstream       *UpstreamEvidence `json:"upstream,omitempty"`
	Link           *ExplanationLink  `json:"link,omitempty"`
	NoLink         *NoLink           `json:"no_link,omitempty"`
	Diagnostics    []Diagnostic      `json:"diagnostics,omitempty"`
}

type ExplanationLink struct {
	CodeBinding   CodeBindingSummary `json:"code_binding"`
	SemanticOwner string             `json:"semantic_owner"`
	Term          TermSummary        `json:"term"`
	OriginPath    PathSummary        `json:"origin_path"`
	Receipt       ReceiptSummary     `json:"receipt"`
	Verifier      VerifierSummary    `json:"verifier"`
}

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
