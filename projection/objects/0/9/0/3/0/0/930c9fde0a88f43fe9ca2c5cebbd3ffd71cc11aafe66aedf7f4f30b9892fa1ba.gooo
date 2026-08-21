package provenance

type lineError struct {
	kind string
	err  error
}

func (e *lineError) Error() string { return e.err.Error() }
func (e *lineError) Unwrap() error { return e.err }

type wireEvidence struct {
	Schema         int               `json:"schema"`
	ID             string            `json:"id"`
	EventID        string            `json:"event_id"`
	SemanticID     string            `json:"semantic_id"`
	Producer       string            `json:"producer"`
	Kind           EvidenceKind      `json:"kind"`
	Status         EvidenceStatus    `json:"status"`
	SourceSpan     wireSourceSpan    `json:"source_span"`
	SourceDigest   string            `json:"source_digest"`
	SemanticDigest string            `json:"semantic_digest"`
	GraphDigest    string            `json:"graph_digest"`
	Sequence       uint64            `json:"sequence"`
	Predecessor    *DigestLink       `json:"predecessor"`
	Attributes     map[string]string `json:"attributes"`
	Freshness      Freshness         `json:"freshness"`
	Hash           string            `json:"hash"`
	Type           string            `json:"type"`
	Subject        string            `json:"subject"`
	GeneratedBy    string            `json:"generated_by"`
}
type wireSourceSpan struct {
	URI   string   `json:"uri"`
	File  string   `json:"file"`
	Start Position `json:"start"`
	End   Position `json:"end"`
}
type canonicalEvidence struct {
	Schema         int                `json:"schema"`
	ID             string             `json:"id"`
	SemanticID     string             `json:"semantic_id"`
	Producer       string             `json:"producer"`
	Kind           EvidenceKind       `json:"kind"`
	Status         EvidenceStatus     `json:"status"`
	SourceSpan     SourceSpan         `json:"source_span"`
	SourceDigest   string             `json:"source_digest"`
	SemanticDigest string             `json:"semantic_digest"`
	GraphDigest    string             `json:"graph_digest"`
	Sequence       uint64             `json:"sequence"`
	Predecessor    *DigestLink        `json:"predecessor,omitempty"`
	Attributes     map[string]string  `json:"attributes,omitempty"`
	Freshness      canonicalFreshness `json:"freshness"`
	Hash           string             `json:"hash,omitempty"`
}
type canonicalFreshness struct {
	ProducedAt string `json:"produced_at"`
	ValidUntil string `json:"valid_until,omitempty"`
}
