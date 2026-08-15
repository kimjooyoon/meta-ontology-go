// Package query provides deterministic, in-memory queries over stable semantic
// IDs and a small PROV relation vocabulary.
//
// The graph is intentionally a projection boundary: it stores relation facts,
// not display names or parser-specific nodes. Deterministic and candidate facts
// remain separate so callers can choose whether a result is authoritative or
// needs review.
package query
