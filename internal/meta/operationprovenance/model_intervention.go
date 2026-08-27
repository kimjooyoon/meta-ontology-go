package operationprovenance

type InterventionResult struct {
	ID                      string `json:"id"`
	Kind                    string `json:"kind"`
	RawSourceDigest         string `json:"raw_source_digest"`
	CanonicalSemanticDigest string `json:"canonical_semantic_digest"`
	ReceiptDigest           string `json:"receipt_digest"`
	DecisionFingerprint     string `json:"decision_fingerprint"`
	TransitionFingerprint   string `json:"transition_fingerprint"`
	RawSourceDigestChanged  bool   `json:"raw_source_digest_changed"`
	SemanticDigestChanged   bool   `json:"semantic_digest_changed"`
	DecisionChanged         bool   `json:"decision_changed"`
	TransitionChanged       bool   `json:"transition_changed"`
	ObservedFailure         bool   `json:"observed_failure"`
	FailureStage            string `json:"failure_stage,omitempty"`
	FailureStep             string `json:"failure_step,omitempty"`
	FailureReason           string `json:"failure_reason,omitempty"`
	Status                  string `json:"status"`
}

type InterventionReport struct {
	Schema      string             `json:"schema"`
	Base        InterventionResult `json:"base"`
	Semantic    InterventionResult `json:"semantic"`
	Nonsemantic InterventionResult `json:"nonsemantic"`
	Noop        InterventionResult `json:"noop"`
	Digest      string             `json:"digest"`
}
