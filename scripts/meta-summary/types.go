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

type sourceThresholds struct {
	GoFile   int `json:"go_file"`
	GoooFile int `json:"gooo_file"`
	Function int `json:"function"`
}

type sourceCandidate struct {
	MetricID  string `json:"metric_id"`
	Subject   string `json:"subject"`
	Actual    int    `json:"actual"`
	Threshold int    `json:"threshold"`
	Role      string `json:"role"`
}

type sourceReadmeObservation struct {
	Applicability string `json:"applicability"`
	Reason        string `json:"reason"`
	Blocking      bool   `json:"blocking"`
}

type selectedSubject struct {
	Operation string `json:"operation"`
	MetricID  string `json:"metric_id"`
	Subject   string `json:"subject"`
}

type sourceInventory struct {
	RegularFiles             int                     `json:"regular_files"`
	DirectoriesIncludingRoot int                     `json:"directories_including_root"`
	DescendantDirectories    int                     `json:"descendant_directories"`
	GoFiles                  int                     `json:"go_files"`
	GoLines                  int                     `json:"go_lines"`
	GoooFiles                int                     `json:"gooo_files"`
	GoooLines                int                     `json:"gooo_lines"`
	RootReadme               sourceReadmeObservation `json:"root_readme"`
	Thresholds               sourceThresholds        `json:"thresholds"`
	Candidates               []sourceCandidate       `json:"over_threshold_candidates"`
	SelectedOperations       int                     `json:"selected_operations"`
	SelectedSubjects         []selectedSubject       `json:"selected_subjects"`
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
	SourceMetrics    sourceInventory    `json:"source_metrics_inventory"`
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
