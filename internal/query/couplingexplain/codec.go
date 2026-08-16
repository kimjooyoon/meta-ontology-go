package couplingexplain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func DecodeVerifiedEnvelope(data []byte) (VerifiedEnvelope, error) {
	var envelope VerifiedEnvelope
	if len(bytes.TrimSpace(data)) == 0 {
		return envelope, fmt.Errorf("verified envelope: empty JSON document")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return VerifiedEnvelope{}, fmt.Errorf("verified envelope: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return VerifiedEnvelope{}, fmt.Errorf("verified envelope: trailing JSON value")
		}
		return VerifiedEnvelope{}, fmt.Errorf("verified envelope: trailing JSON: %w", err)
	}
	return envelope, nil
}

func (e VerifiedEnvelope) CanonicalJSON() ([]byte, error) {
	diagnostics := canonicalDiagnostics(e.Diagnostics)
	return json.Marshal(canonicalEnvelope{
		Schema: e.Schema, Binding: e.Binding,
		Upstream:    toCanonicalUpstream(e.Upstream),
		CodeBinding: toCanonicalCodeBinding(e.CodeBinding), SemanticOwner: e.SemanticOwner,
		Term: toCanonicalTerm(e.Term), OriginPath: toCanonicalPath(e.OriginPath),
		Receipt: toCanonicalReceipt(e.Receipt), Verifier: toCanonicalVerifier(e.Verifier),
		Verdict: e.Verdict, NoLinkReason: e.NoLinkReason,
		EvidenceDigest: e.EvidenceDigest, Diagnostics: diagnostics,
	})
}

// Digest is the exact envelope digest expected in Request. EnvelopeDigest is
// excluded to avoid a self-referential hash.
func (e VerifiedEnvelope) Digest() string {
	data, err := e.CanonicalJSON()
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

func (e Explanation) CanonicalJSON(view View) ([]byte, error) {
	if view != ViewCompact && view != ViewExpanded {
		return nil, fmt.Errorf("unknown explanation view %q", view)
	}
	value := canonicalExplanation{
		Status: e.Status, EvidenceDigest: e.EvidenceDigest, Binding: e.Binding,
		Upstream:    toCanonicalUpstream(e.Upstream),
		Diagnostics: canonicalDiagnostics(e.Diagnostics),
	}
	if e.NoLink != nil {
		value.NoLink = &NoLink{Reason: e.NoLink.Reason}
	}
	if e.Link != nil {
		value.Link = toCanonicalLink(*e.Link, view == ViewExpanded)
	}
	return json.Marshal(value)
}

func (e Explanation) CanonicalDigest(view View) (string, error) {
	data, err := e.CanonicalJSON(view)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func DigestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func (p *PathStep) UnmarshalJSON(data []byte) error {
	var wire struct {
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
	if err := strictDecode(data, &wire); err != nil {
		return err
	}
	p.FromID, p.ToID, p.Kind = wire.FromID, wire.ToID, wire.Kind
	p.Phase = semantic.PhasePlacement{Phase: wire.Phase, Ordinal: wire.PhaseOrdinal}
	p.RuleRef, p.InputDigest, p.OutputDigest, p.EvidenceRef = wire.RuleRef, wire.InputDigest, wire.OutputDigest, wire.EvidenceRef
	return nil
}

func (p PathStep) MarshalJSON() ([]byte, error) {
	return json.Marshal(canonicalPathStep{
		FromID: p.FromID, ToID: p.ToID, Kind: p.Kind, Phase: p.Phase.Phase,
		PhaseOrdinal: p.Phase.Ordinal, RuleRef: p.RuleRef,
		InputDigest: p.InputDigest, OutputDigest: p.OutputDigest, EvidenceRef: p.EvidenceRef,
	})
}

func strictDecode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}

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
}

func toCanonicalCodeBinding(value CodeBindingSummary) canonicalCodeBinding {
	return canonicalCodeBinding{CodeSymbolID: value.CodeSymbolID, SemanticOwnerID: value.SemanticOwnerID,
		RegisteredSurfaceID: value.RegisteredSurfaceID, SourceMapID: value.SourceMapID, BindingDigest: value.BindingDigest}
}

type canonicalTerm struct {
	TermID           string `json:"term_id"`
	SemanticOwnerID  string `json:"semantic_owner_id"`
	Version          string `json:"version"`
	DefinitionDigest string `json:"definition_digest"`
}

func toCanonicalTerm(value TermSummary) canonicalTerm {
	return canonicalTerm{TermID: value.TermID, SemanticOwnerID: value.SemanticOwnerID,
		Version: value.Version, DefinitionDigest: value.DefinitionDigest}
}

type canonicalPath struct {
	PathID     string              `json:"path_id"`
	StartID    string              `json:"start_id"`
	EndID      string              `json:"end_id"`
	StepCount  int                 `json:"step_count"`
	PathDigest string              `json:"path_digest"`
	Steps      []canonicalPathStep `json:"steps,omitempty"`
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
	DeltaDigest    string                      `json:"delta_digest,omitempty"`
	OriginPathID   string                      `json:"origin_path_id"`
	EvidenceRefs   []string                    `json:"evidence_refs"`
}

func toCanonicalReceipt(value ReceiptSummary) canonicalReceipt {
	return canonicalReceipt{ReceiptID: value.ReceiptID, SurfaceID: value.SurfaceID, ChangeClaim: value.ChangeClaim,
		ReceiptKind: value.ReceiptKind, BeforeIRDigest: value.BeforeIRDigest, AfterIRDigest: value.AfterIRDigest,
		DeltaDigest: value.DeltaDigest, OriginPathID: value.OriginPathID, EvidenceRefs: sortedStrings(value.EvidenceRefs)}
}

type canonicalVerifier struct {
	EvidenceID     string        `json:"evidence_id"`
	ReceiptID      string        `json:"receipt_id"`
	State          VerifierState `json:"state"`
	Independent    bool          `json:"independent"`
	EvidenceDigest string        `json:"evidence_digest"`
	EvidenceRefs   []string      `json:"evidence_refs"`
}

func toCanonicalVerifier(value VerifierSummary) canonicalVerifier {
	return canonicalVerifier{EvidenceID: value.EvidenceID, ReceiptID: value.ReceiptID, State: value.State,
		Independent: value.Independent, EvidenceDigest: value.EvidenceDigest, EvidenceRefs: sortedStrings(value.EvidenceRefs)}
}

type canonicalExplanation struct {
	Status         Status             `json:"status"`
	EvidenceDigest string             `json:"evidence_digest"`
	Binding        SnapshotBinding    `json:"binding"`
	Upstream       *canonicalUpstream `json:"upstream,omitempty"`
	Link           *canonicalLink     `json:"link,omitempty"`
	NoLink         *NoLink            `json:"no_link,omitempty"`
	Diagnostics    []Diagnostic       `json:"diagnostics,omitempty"`
}

type canonicalLink struct {
	CodeBinding   canonicalCodeBinding `json:"code_binding"`
	SemanticOwner string               `json:"semantic_owner"`
	Term          canonicalTerm        `json:"term"`
	OriginPath    canonicalPath        `json:"origin_path"`
	Receipt       canonicalReceipt     `json:"receipt"`
	Verifier      canonicalVerifier    `json:"verifier"`
}

type canonicalUpstream struct {
	SourceMapDigest      string            `json:"source_map_digest,omitempty"`
	ManifestDigest       string            `json:"manifest_digest"`
	DetectorInputDigest  string            `json:"detector_input_digest"`
	DetectorResultDigest string            `json:"detector_result_digest"`
	DetectorStatus       detector.Status   `json:"detector_status"`
	DetectorReasons      []detector.Reason `json:"detector_reasons,omitempty"`
	ManifestStatus       string            `json:"manifest_status,omitempty"`
	ManifestReason       string            `json:"manifest_reason,omitempty"`
}

func toCanonicalUpstream(value *UpstreamEvidence) *canonicalUpstream {
	if value == nil {
		return nil
	}
	reasons := append([]detector.Reason(nil), value.DetectorReasons...)
	sort.Slice(reasons, func(i, j int) bool {
		if reasons[i].Code != reasons[j].Code {
			return reasons[i].Code < reasons[j].Code
		}
		return reasons[i].Detail < reasons[j].Detail
	})
	return &canonicalUpstream{
		SourceMapDigest: value.SourceMapDigest, ManifestDigest: value.ManifestDigest,
		DetectorInputDigest: value.DetectorInputDigest, DetectorResultDigest: value.DetectorResultDigest,
		DetectorStatus: value.DetectorStatus, DetectorReasons: reasons,
		ManifestStatus: string(value.ManifestStatus), ManifestReason: string(value.ManifestReason),
	}
}

func toCanonicalLink(value ExplanationLink, expanded bool) *canonicalLink {
	path := toCanonicalPath(value.OriginPath)
	if !expanded {
		path.Steps = nil
	}
	return &canonicalLink{CodeBinding: toCanonicalCodeBinding(value.CodeBinding), SemanticOwner: value.SemanticOwner,
		Term: toCanonicalTerm(value.Term), OriginPath: path, Receipt: toCanonicalReceipt(value.Receipt), Verifier: toCanonicalVerifier(value.Verifier)}
}

func canonicalDiagnostics(values []Diagnostic) []Diagnostic {
	diagnostics := make([]Diagnostic, 0, len(values))
	for _, value := range values {
		diagnostics = append(diagnostics, Diagnostic{Code: value.Code, IDs: sortedStrings(value.IDs)})
	}
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Code != diagnostics[j].Code {
			return diagnostics[i].Code < diagnostics[j].Code
		}
		return joinIDs(diagnostics[i].IDs) < joinIDs(diagnostics[j].IDs)
	})
	return diagnostics
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func joinIDs(values []string) string {
	result := ""
	for _, value := range values {
		result += value + "\x00"
	}
	return result
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
