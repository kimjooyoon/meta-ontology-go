package main

const (
	summarySchema     = "gooo/ci-summary-projection/v1"
	defaultLimitBytes = 900 * 1024
)

type options struct {
	MetricsPath    string
	PlanPath       string
	ExecutionPath  string
	ReceiptsPath   string
	ProvenancePath string
	OutputPath     string
	ReportPath     string
	LimitBytes     int
}

type artifactEvidence struct {
	ID     string `json:"id"`
	File   string `json:"file"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type metricIndicator struct {
	ID       string `json:"id"`
	Route    string `json:"route"`
	Verdict  string `json:"verdict"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}

type summaryReport struct {
	SchemaVersion    string             `json:"schema_version"`
	Decision         string             `json:"decision"`
	Reason           string             `json:"reason"`
	InputSHA256      string             `json:"input_sha256"`
	OutputSHA256     string             `json:"output_sha256"`
	OutputBytes      int                `json:"output_bytes"`
	LimitBytes       int                `json:"limit_bytes"`
	ProvenanceSchema string             `json:"provenance_schema"`
	Provenance       provenanceEvidence `json:"provenance"`
	Indicators       []metricIndicator  `json:"indicators"`
	Artifacts        []artifactEvidence `json:"artifacts"`
}

type provenanceEvidence struct {
	Decision     string `json:"decision"`
	Reason       string `json:"reason"`
	LedgerDigest string `json:"ledger_digest"`
	LedgerCount  int    `json:"ledger_count"`
	Pass         int    `json:"pass"`
	Envelope     string `json:"envelope_digest"`
	Replay       string `json:"replay_digest"`
}
