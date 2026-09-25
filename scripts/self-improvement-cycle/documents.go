package main

type metricsPolicy struct {
	ExemptProjectRootTopology bool `json:"exempt_project_root_topology"`
	ExemptProjectRootREADME   bool `json:"exempt_project_root_readme"`
}

type metricsIndicator struct {
	MetricID            string `json:"metric_id"`
	Subject             string `json:"subject"`
	Value               int    `json:"value"`
	Applicability       string `json:"applicability"`
	ApplicabilityReason string `json:"applicability_reason"`
	Blocking            bool   `json:"blocking"`
	Decision            string `json:"decision"`
	Detail              string `json:"detail"`
}

type metricsMeta struct {
	Schema     string             `json:"schema"`
	Policy     metricsPolicy      `json:"policy"`
	Indicators []metricsIndicator `json:"indicators"`
}

type metricsDocument struct {
	CommitSHA          string            `json:"commit_sha"`
	Files              []metricsFile     `json:"files"`
	Meta               metricsMeta       `json:"meta"`
	Directories        []MetricsSnapshot `json:"directories"`
	StorageDirectories []MetricsSnapshot `json:"storage_directories"`
}

type executionDocument struct {
	SchemaVersion                 string `json:"schema_version"`
	BaseSHA                       string `json:"base_sha"`
	HeadSHA                       string `json:"head_sha"`
	PlanDigest                    string `json:"plan_digest"`
	ManifestDigest                string `json:"manifest_digest"`
	IndicatorDecisionLedgerDigest string `json:"indicator_decision_ledger_digest"`
	IndicatorDecisionLedgerCount  int    `json:"indicator_decision_ledger_count"`
	Decision                      string `json:"decision"`
	Reason                        string `json:"reason"`
	PromotionAuthorized           bool   `json:"promotion_authorized"`
}

type receiptDocument struct {
	SchemaVersion                 string `json:"schema_version"`
	BaseSHA                       string `json:"base_sha"`
	HeadSHA                       string `json:"head_sha"`
	PlanDigest                    string `json:"plan_digest"`
	ReportDigest                  string `json:"report_digest"`
	IndicatorDecisionLedgerDigest string `json:"indicator_decision_ledger_digest"`
	IndicatorDecisionLedgerCount  int    `json:"indicator_decision_ledger_count"`
	Decision                      string `json:"decision"`
	Reason                        string `json:"reason"`
	PromotionAuthorized           bool   `json:"promotion_authorized"`
}

type provenanceDocument struct {
	SchemaVersion                 string `json:"schema_version"`
	BaseSHA                       string `json:"base_sha"`
	HeadSHA                       string `json:"head_sha"`
	PlanDigest                    string `json:"plan_digest"`
	ExecutionManifestDigest       string `json:"execution_manifest_digest"`
	ReceiptReportDigest           string `json:"receipt_report_digest"`
	IndicatorDecisionLedgerDigest string `json:"indicator_decision_ledger_digest"`
	IndicatorDecisionLedgerCount  int    `json:"indicator_decision_ledger_count"`
	Decision                      string `json:"decision"`
	Reason                        string `json:"reason"`
	PromotionAuthorized           bool   `json:"promotion_authorized"`
	EnvelopeDigest                string `json:"envelope_digest"`
}
