package couplingexplain

import (
	"encoding/json"
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
