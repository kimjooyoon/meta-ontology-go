package couplingexplain

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type canonicalEnvelope struct {
	Schema         string               `json:"schema"`
	Binding        SnapshotBinding      `json:"binding"`
	Upstream       *canonicalUpstream   `json:"upstream,omitempty"`
	CodeBinding    canonicalCodeBinding `json:"code_binding"`
	SemanticOwner  string               `json:"semantic_owner"`
	Term           canonicalTerm        `json:"term"`
	OriginPath     canonicalPath        `json:"origin_path"`
	Receipt        canonicalReceipt     `json:"receipt"`
	Verifier       canonicalVerifier    `json:"verifier"`
	Verdict        EnvelopeVerdict      `json:"verdict"`
	NoLinkReason   LinkReason           `json:"no_link_reason,omitempty"`
	EvidenceDigest string               `json:"evidence_digest"`
	Diagnostics    []Diagnostic         `json:"diagnostics,omitempty"`
}

type canonicalCodeBinding struct {
	CodeSymbolID        string `json:"code_symbol_id"`
	SemanticOwnerID     string `json:"semantic_owner_id"`
	RegisteredSurfaceID string `json:"registered_surface_id"`
	SourceMapID         string `json:"source_map_id"`
	BindingDigest       string `json:"binding_digest"`
	CodeBindingDigest   string `json:"code_binding_digest"`
}

type canonicalCodeBindingUnsigned struct {
	CodeSymbolID        string `json:"code_symbol_id"`
	SemanticOwnerID     string `json:"semantic_owner_id"`
	RegisteredSurfaceID string `json:"registered_surface_id"`
	SourceMapID         string `json:"source_map_id"`
	BindingDigest       string `json:"binding_digest"`
}

func toCanonicalCodeBinding(value CodeBindingSummary) canonicalCodeBinding {
	return canonicalCodeBinding{CodeSymbolID: value.CodeSymbolID, SemanticOwnerID: value.SemanticOwnerID,
		RegisteredSurfaceID: value.RegisteredSurfaceID, SourceMapID: value.SourceMapID, BindingDigest: value.BindingDigest,
		CodeBindingDigest: value.CodeBindingDigest}
}

func codeBindingDigest(value CodeBindingSummary) string {
	data, err := json.Marshal(canonicalCodeBindingUnsigned{CodeSymbolID: value.CodeSymbolID, SemanticOwnerID: value.SemanticOwnerID,
		RegisteredSurfaceID: value.RegisteredSurfaceID, SourceMapID: value.SourceMapID, BindingDigest: value.BindingDigest})
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

type canonicalTerm struct {
	TermID           string `json:"term_id"`
	SemanticOwnerID  string `json:"semantic_owner_id"`
	Version          string `json:"version"`
	DefinitionDigest string `json:"definition_digest"`
}

type canonicalTermUnsigned struct {
	TermID          string `json:"term_id"`
	SemanticOwnerID string `json:"semantic_owner_id"`
	Version         string `json:"version"`
}

func toCanonicalTerm(value TermSummary) canonicalTerm {
	return canonicalTerm{TermID: value.TermID, SemanticOwnerID: value.SemanticOwnerID,
		Version: value.Version, DefinitionDigest: value.DefinitionDigest}
}

func termDefinitionDigest(value TermSummary) string {
	data, err := json.Marshal(canonicalTermUnsigned{TermID: value.TermID, SemanticOwnerID: value.SemanticOwnerID, Version: value.Version})
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

type canonicalPath struct {
	PathID     string              `json:"path_id"`
	StartID    string              `json:"start_id"`
	EndID      string              `json:"end_id"`
	StepCount  int                 `json:"step_count"`
	PathDigest string              `json:"path_digest"`
	Steps      []canonicalPathStep `json:"steps,omitempty"`
}

type canonicalPathUnsigned struct {
	PathID    string              `json:"path_id"`
	StartID   string              `json:"start_id"`
	EndID     string              `json:"end_id"`
	StepCount int                 `json:"step_count"`
	Steps     []canonicalPathStep `json:"steps,omitempty"`
}

func toCanonicalPath(value PathSummary) canonicalPath {
	steps := make([]canonicalPathStep, 0, len(value.Steps))
	for _, step := range value.Steps {
		steps = append(steps, canonicalPathStep{FromID: step.FromID, ToID: step.ToID, Kind: step.Kind,
			Phase: step.Phase.Phase, PhaseOrdinal: step.Phase.Ordinal, RuleRef: step.RuleRef,
			InputDigest: step.InputDigest, OutputDigest: step.OutputDigest, EvidenceRef: step.EvidenceRef})
	}
	return canonicalPath{PathID: value.PathID, StartID: value.StartID, EndID: value.EndID,
		StepCount: value.StepCount, PathDigest: value.PathDigest, Steps: steps}
}

func pathDigest(value PathSummary) string {
	path := toCanonicalPath(value)
	data, err := json.Marshal(canonicalPathUnsigned{PathID: path.PathID, StartID: path.StartID, EndID: path.EndID,
		StepCount: path.StepCount, Steps: path.Steps})
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

type canonicalPathStep struct {
	FromID       string                  `json:"from_id"`
	ToID         string                  `json:"to_id"`
	Kind         semantic.InferenceKind  `json:"kind"`
	Phase        semantic.InferencePhase `json:"phase"`
	PhaseOrdinal uint64                  `json:"phase_ordinal"`
	RuleRef      string                  `json:"rule_ref,omitempty"`
	InputDigest  string                  `json:"input_digest"`
	OutputDigest string                  `json:"output_digest"`
	EvidenceRef  string                  `json:"evidence_ref,omitempty"`
}

type canonicalReceipt struct {
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
}

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
