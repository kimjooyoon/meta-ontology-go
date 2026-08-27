package operationprovenance

type InterventionResult struct {
	ID                      string `json:"id"`
	Kind                    string `json:"kind"`
	RawSourceDigest         string `json:"raw_source_digest"`
	CanonicalSemanticDigest string `json:"canonical_semantic_digest"`
	ReceiptDigest           string `json:"receipt_digest"`
	DecisionFingerprint     string `json:"decision_fingerprint"`
	TransitionFingerprint   string `json:"transition_fingerprint"`
	SemanticDigestChanged   bool   `json:"semantic_digest_changed"`
	DecisionChanged         bool   `json:"decision_changed"`
	TransitionChanged       bool   `json:"transition_changed"`
	Status                  string `json:"status"`
}

type InterventionReport struct {
	Schema      string             `json:"schema"`
	Base        InterventionResult `json:"base"`
	Semantic    InterventionResult `json:"semantic"`
	Nonsemantic InterventionResult `json:"nonsemantic"`
	Digest      string             `json:"digest"`
}
