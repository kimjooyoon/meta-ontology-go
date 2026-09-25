package query

// EnvelopeError is a stable machine-readable rejection. Message is diagnostic
// context; Code is the compatibility key for adapters and tests.
type EnvelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	cause   error  `json:"-"`
}

func (queryError *EnvelopeError) Error() string {
	if queryError == nil {
		return ""
	}
	return queryError.Code + ": " + queryError.Message
}
func (queryError *EnvelopeError) Unwrap() error {
	if queryError == nil {
		return nil
	}
	return queryError.cause
}

// EnvelopeMetadata exposes projection and authority state without promoting a
// query graph or fabricating missing source/provenance evidence.
type EnvelopeMetadata struct {
	SchemaVersion     string           `json:"schema_version"`
	Namespace         string           `json:"namespace,omitempty"`
	GraphHash         string           `json:"graph_hash"`
	SemanticDigest    string           `json:"semantic_digest,omitempty"`
	ProjectionStatus  string           `json:"projection_status"`
	SourceStatus      string           `json:"source_status"`
	IRStatus          string           `json:"ir_status"`
	EvidenceStatus    string           `json:"evidence_status"`
	ProvenanceStatus  string           `json:"provenance_status"`
	DerivedStatus     string           `json:"derived_status"`
	DerivedRuleSchema string           `json:"derived_rule_schema,omitempty"`
	DerivedRuleDigest string           `json:"derived_rule_digest,omitempty"`
	AuthorityLabels   []AuthorityLabel `json:"authority_labels"`
}

// Response is the versioned machine-readable result envelope. Hash is the
// digest of CanonicalJSON with Hash omitted; it is a view receipt, not SSOT.
type Response struct {
	Schema      string           `json:"schema"`
	Status      string           `json:"status"`
	Request     Request          `json:"request"`
	RequestHash string           `json:"request_hash,omitempty"`
	Result      QueryResult      `json:"result"`
	Metadata    EnvelopeMetadata `json:"metadata"`
	Error       *EnvelopeError   `json:"error,omitempty"`
	Hash        string           `json:"canonical_hash,omitempty"`
}

const (
	ResponseOK     = "ok"
	ResponseError  = "error"
	StatusDeferred = "deferred"
)
