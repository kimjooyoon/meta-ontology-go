package couplingexplain

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
	Presentation   Presentation                `json:"presentation"`
}
type VerifierSummary struct {
	EvidenceID     string        `json:"evidence_id"`
	ReceiptID      string        `json:"receipt_id"`
	State          VerifierState `json:"state"`
	Independent    bool          `json:"independent"`
	EvidenceDigest string        `json:"evidence_digest"`
	VerifierDigest string        `json:"verifier_digest"`
	EvidenceRefs   []string      `json:"evidence_refs"`
	Presentation   Presentation  `json:"presentation"`
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
