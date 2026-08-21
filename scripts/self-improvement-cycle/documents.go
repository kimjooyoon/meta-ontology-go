package main

type metricsDocument struct {
	CommitSHA string `json:"commit_sha"`
}

type planDocument struct {
	SchemaVersion       string `json:"schema_version"`
	BaseSHA             string `json:"base_sha"`
	HeadSHA             string `json:"head_sha"`
	PlanDigest          string `json:"plan_digest"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
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

type contractIndicator struct {
	Route   string `json:"route"`
	Verdict string `json:"verdict"`
}

type contractCoverage struct {
	Covered bool `json:"covered"`
}

type contractDocument struct {
	Schema              string              `json:"schema"`
	CommitSHA           string              `json:"commit_sha"`
	SourceSHA256        string              `json:"source_sha256"`
	SemanticHash        string              `json:"semantic_hash"`
	RegistryDigest      string              `json:"registry_digest"`
	Status              string              `json:"status"`
	PromotionAuthorized bool                `json:"promotion_authorized"`
	Indicators          []contractIndicator `json:"indicators"`
	ExecutorCoverage    []contractCoverage  `json:"executor_coverage"`
}
