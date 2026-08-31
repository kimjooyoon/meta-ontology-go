package assuranceeligibility

type assuranceReport struct {
	Schema            string `json:"schema"`
	SubjectSHA        string `json:"subject_sha"`
	DenominatorID     string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	ReportDigest      string `json:"report_digest"`
	Summary           struct {
		DenominatorTotal           int `json:"denominator_total"`
		Operating                  int `json:"operating"`
		NotImplemented             int `json:"not_implemented"`
		ImplementationCoverageBPS  int `json:"implementation_coverage_bps"`
		EvidenceGroupsObserved     int `json:"evidence_groups_observed"`
		EvidenceGroupsTotal        int `json:"evidence_groups_total"`
		UnknownTopDecisions        int `json:"unknown_top_decisions"`
		SnapshotBindingsObserved   int `json:"snapshot_bindings_observed"`
		SnapshotBindingsRequired   int `json:"snapshot_bindings_required"`
		RawReconstructionsObserved int `json:"raw_reconstructions_observed"`
		RawReconstructionsRequired int `json:"raw_reconstructions_required"`
		UnresolvedIndicators       int `json:"unresolved_indicators"`
		ViolatedGuardrails         int `json:"violated_guardrails"`
		RepositoryWrites           int `json:"repository_writes"`
	} `json:"summary"`
	Obligations []struct {
		MetricID   string `json:"metric_id"`
		Status     string `json:"status"`
		Resolution string `json:"resolution"`
	} `json:"obligations"`
}
