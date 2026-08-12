package query

const QueryProjectionSchemaVersion = "gooo-query/v1"

// AuthorityLabel makes the authority boundary explicit to consumers of a
// query projection. A query graph is never authoritative merely because it
// contains a stable hash.
type AuthorityLabel struct {
	View      string `json:"view"`
	Authority string `json:"authority"`
	Status    string `json:"status"`
}

// ProjectionMetadata binds a query view to the semantic IR snapshot it was
// derived from. The empty SourceDigest is intentional: this package cannot
// invent a source digest that was not supplied by the IR producer. Evidence
// and provenance digests come from the validated IR; their status distinguishes
// an available set from a known-empty set.
type ProjectionMetadata struct {
	SchemaVersion    string
	GraphHash        string
	SemanticDigest   string
	SourceDigest     string
	EvidenceDigest   string
	ProvenanceDigest string
	SourceStatus     string
	EvidenceStatus   string
	ProvenanceStatus string
	ProjectionStatus string
	AuthorityLabels  []AuthorityLabel
}
