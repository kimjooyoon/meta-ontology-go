package main

type provenanceIndicator struct {
	ID             string `json:"id"`
	Route          string `json:"route"`
	Verdict        string `json:"verdict"`
	EvidenceDigest string `json:"evidence_digest"`
}

func (indicator provenanceIndicator) valid() bool {
	route := indicator.Route == "FOUNDATION" || indicator.Route == "COHERENCE" || indicator.Route == "REGRESSION"
	return indicator.ID != "" && route && indicator.Verdict == "PASS" && hexDigest(indicator.EvidenceDigest, 64)
}

type provenanceSummary struct {
	Pass    int `json:"pass"`
	Fail    int `json:"fail"`
	Unknown int `json:"unknown"`
}

type provenanceEnvelope struct {
	SchemaVersion                 string                `json:"schema_version"`
	BaseSHA                       string                `json:"base_sha"`
	HeadSHA                       string                `json:"head_sha"`
	PlanDigest                    string                `json:"plan_digest"`
	ExecutionManifestDigest       string                `json:"execution_manifest_digest"`
	ReceiptReportDigest           string                `json:"receipt_report_digest"`
	IndicatorDecisionLedgerDigest string                `json:"indicator_decision_ledger_digest"`
	IndicatorDecisionLedgerCount  int                   `json:"indicator_decision_ledger_count"`
	InputDigest                   string                `json:"input_digest"`
	Decision                      string                `json:"decision"`
	Reason                        string                `json:"reason"`
	Indicators                    []provenanceIndicator `json:"indicators"`
	Summary                       provenanceSummary     `json:"summary"`
	PromotionAuthorized           bool                  `json:"promotion_authorized"`
	EnvelopeDigest                string                `json:"envelope_digest"`
	ReplayDigest                  string                `json:"replay_digest"`
}
