package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// InferenceQueryResult is a canonical, non-authoritative query receipt.
// Complete is false for every rejection or budget overrun; rejected results
// contain no partial rows.
type InferenceQueryResult struct {
	Schema      string                 `json:"schema"`
	Status      string                 `json:"status"`
	Request     InferenceQuery         `json:"request"`
	RequestHash string                 `json:"request_hash,omitempty"`
	Edges       []InferenceRow         `json:"edges,omitempty"`
	Claims      []SemanticChangeRow    `json:"claims,omitempty"`
	Evidence    []InferenceEvidenceRow `json:"evidence,omitempty"`
	Chain       *InferenceChainResult  `json:"chain,omitempty"`
	Work        InferenceWork          `json:"work"`
	Complete    bool                   `json:"complete"`
	Error       *EnvelopeError         `json:"error,omitempty"`
	Hash        string                 `json:"canonical_hash,omitempty"`
}

// InferenceResponse and InferenceResult are compatibility spellings for
// callers that use the existing envelope/result terminology.
type InferenceResponse = InferenceQueryResult
type InferenceResult = InferenceQueryResult

// InferenceProjection is a detached, read-only view of one validated and
// normalized InferencePathV1 snapshot.
type InferenceProjection struct {
	path semantic.InferencePathV1
}
