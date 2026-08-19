package couplingexplain

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type canonicalReceiptUnsigned struct {
	ReceiptID      string                      `json:"receipt_id"`
	SurfaceID      string                      `json:"surface_id"`
	ChangeClaim    ChangeClaim                 `json:"change_claim"`
	ReceiptKind    semantic.SemanticChangeKind `json:"receipt_kind"`
	BeforeIRDigest string                      `json:"before_ir_digest"`
	AfterIRDigest  string                      `json:"after_ir_digest"`
	CanonicalDelta string                      `json:"canonical_delta,omitempty"`
	DeltaDigest    string                      `json:"delta_digest,omitempty"`
	OriginPathID   string                      `json:"origin_path_id"`
	EvidenceRefs   []string                    `json:"evidence_refs"`
}

func toCanonicalReceipt(value ReceiptSummary) canonicalReceipt {
	return canonicalReceipt{ReceiptID: value.ReceiptID, SurfaceID: value.SurfaceID, ChangeClaim: value.ChangeClaim,
		ReceiptKind: value.ReceiptKind, BeforeIRDigest: value.BeforeIRDigest, AfterIRDigest: value.AfterIRDigest,
		CanonicalDelta: value.CanonicalDelta, DeltaDigest: value.DeltaDigest, ReceiptDigest: value.ReceiptDigest,
		OriginPathID: value.OriginPathID, EvidenceRefs: sortedStrings(value.EvidenceRefs)}
}
func receiptDigest(value ReceiptSummary) string {
	data, err := json.Marshal(canonicalReceiptUnsigned{ReceiptID: value.ReceiptID, SurfaceID: value.SurfaceID,
		ChangeClaim: value.ChangeClaim, ReceiptKind: value.ReceiptKind, BeforeIRDigest: value.BeforeIRDigest,
		AfterIRDigest: value.AfterIRDigest, CanonicalDelta: value.CanonicalDelta, DeltaDigest: value.DeltaDigest,
		OriginPathID: value.OriginPathID, EvidenceRefs: sortedStrings(value.EvidenceRefs)})
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

type canonicalVerifier struct {
	EvidenceID     string        `json:"evidence_id"`
	ReceiptID      string        `json:"receipt_id"`
	State          VerifierState `json:"state"`
	Independent    bool          `json:"independent"`
	EvidenceDigest string        `json:"evidence_digest"`
	VerifierDigest string        `json:"verifier_digest"`
	EvidenceRefs   []string      `json:"evidence_refs"`
}
type canonicalVerifierUnsigned struct {
	EvidenceID     string        `json:"evidence_id"`
	ReceiptID      string        `json:"receipt_id"`
	State          VerifierState `json:"state"`
	Independent    bool          `json:"independent"`
	EvidenceDigest string        `json:"evidence_digest"`
	EvidenceRefs   []string      `json:"evidence_refs"`
}

func toCanonicalVerifier(value VerifierSummary) canonicalVerifier {
	return canonicalVerifier{EvidenceID: value.EvidenceID, ReceiptID: value.ReceiptID, State: value.State,
		Independent: value.Independent, EvidenceDigest: value.EvidenceDigest, VerifierDigest: value.VerifierDigest,
		EvidenceRefs: sortedStrings(value.EvidenceRefs)}
}
func verifierDigest(value VerifierSummary) string {
	data, err := json.Marshal(canonicalVerifierUnsigned{EvidenceID: value.EvidenceID, ReceiptID: value.ReceiptID,
		State: value.State, Independent: value.Independent, EvidenceDigest: value.EvidenceDigest,
		EvidenceRefs: sortedStrings(value.EvidenceRefs)})
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}
