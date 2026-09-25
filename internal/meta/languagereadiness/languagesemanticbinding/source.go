package languagesemanticbinding

type Source struct {
	ExpectedHeadSHA         string `json:"expected_head_sha"`
	ContractSchema          string `json:"contract_schema"`
	RegistryDigest          string `json:"registry_digest"`
	ReadinessFileDigest     string `json:"readiness_file_digest"`
	ReadinessArtifactDigest string `json:"readiness_artifact_digest"`
	ConceptFileDigest       string `json:"concept_file_digest"`
	ConceptArtifactDigest   string `json:"concept_artifact_digest"`
	ConceptEvidenceDigest   string `json:"concept_evidence_digest"`
	SemanticFileDigest      string `json:"semantic_file_digest"`
	SemanticReportDigest    string `json:"semantic_report_digest"`
	SemanticRegistryDigest  string `json:"semantic_registry_digest"`
	Producer                string `json:"producer"`
	Consumer                string `json:"consumer"`
	MetaOperation           string `json:"meta_operation"`
}
