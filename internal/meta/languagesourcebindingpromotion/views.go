package languagesourcebindingpromotion

type producerCase struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type producerSummary struct {
	CasesSatisfied int `json:"cases_satisfied"`
	CasesTotal     int `json:"cases_total"`
}

type oracleCase struct {
	ID                 string `json:"id"`
	Status             string `json:"status"`
	ObservedDecision   string `json:"observed_decision"`
	ObservedResolution string `json:"observed_resolution"`
	ObservedReason     string `json:"observed_reason"`
	SourceDigest       string `json:"source_digest"`
	ArtifactDigest     string `json:"artifact_digest"`
}

type oracleSummary struct {
	CasesSatisfied           int `json:"cases_satisfied"`
	CasesTotal               int `json:"cases_total"`
	ProducerDependencies     int `json:"producer_dependencies"`
	SemanticCorrectnessClaims int `json:"semantic_correctness_claims"`
}
