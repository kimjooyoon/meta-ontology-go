package transformationeffect

// CausalUnknownRecord is the stable, receipt-derived identity of one unknown
// obligation. It retains the failure context for dependency-blocked records.
type CausalUnknownRecord struct {
	ActionIndicatorID   string   `json:"action_indicator_id"`
	RequiredIndicatorID string   `json:"required_indicator_id"`
	Stage               string   `json:"stage"`
	Step                string   `json:"step"`
	Reason              string   `json:"reason"`
	UnknownClass        string   `json:"unknown_class"`
	NextOperation       string   `json:"next_operation"`
	BlockedBy           []string `json:"blocked_by"`
}

// CausalUnknownProjection is the canonical aggregate consumed by downstream
// verifiers. Digest is excluded from its own input by design.
type CausalUnknownProjection struct {
	DirectUnknownCount            int                   `json:"direct_unknown_count"`
	DependencyBlockedUnknownCount int                   `json:"dependency_blocked_unknown_count"`
	Records                       []CausalUnknownRecord `json:"records"`
	Digest                        string                `json:"-"`
}
