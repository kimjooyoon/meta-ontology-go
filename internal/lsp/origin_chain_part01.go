package lsp

import "github.com/kimjooyoon/meta-ontology-go/internal/provenance"

// OriginChainReference is the non-authorizing LSP-facing projection of a
// source-to-runtime provenance chain. It allows clients to display a stable
// origin without treating the diagnostic surface as execution authority.
type OriginChainReference struct {
	DeclarationURI        string                       `json:"declaration_uri"`
	DeclarationSymbol     string                       `json:"declaration_symbol"`
	IRNode                string                       `json:"ir_node"`
	GeneratedURI          string                       `json:"generated_uri"`
	GeneratedSymbol       string                       `json:"generated_symbol"`
	ReverseObservationURI string                       `json:"reverse_observation_uri"`
	ReverseObservation    string                       `json:"reverse_observation"`
	MetricName            string                       `json:"metric_name"`
	MetricValue           string                       `json:"metric_value"`
	EvidenceDigest        string                       `json:"evidence_digest"`
	Status                provenance.OriginChainStatus `json:"status"`
	Digest                string                       `json:"digest,omitempty"`
	Reason                string                       `json:"reason,omitempty"`
	NonAuthorizing        bool                         `json:"non_authorizing"`
}

// NewOriginChainReference projects an origin observation for LSP clients.
// Only COMPLETE observations are returned as comparable references; partial
// and unknown evidence remains visible but is never promoted to proof.
func NewOriginChainReference(observation provenance.OriginChainObservation) (OriginChainReference, bool) {
	chain := observation.Chain
	reference := OriginChainReference{
		DeclarationURI:        chain.DeclarationURI,
		DeclarationSymbol:     chain.DeclarationSymbol,
		IRNode:                chain.IRNode,
		GeneratedURI:          chain.GeneratedURI,
		GeneratedSymbol:       chain.GeneratedSymbol,
		ReverseObservationURI: chain.ReverseObservationURI,
		ReverseObservation:    chain.ReverseObservation,
		MetricName:            chain.MetricName,
		MetricValue:           chain.MetricValue,
		EvidenceDigest:        chain.EvidenceDigest,
		Status:                observation.Status,
		Digest:                observation.Digest,
		Reason:                observation.Reason,
		NonAuthorizing:        true,
	}
	return reference, observation.Comparable()
}
