package operationprovenance

type SourceReconstruction struct {
	Numerator               int `json:"numerator"`
	Denominator             int `json:"denominator"`
	MetricFieldsNumerator   int `json:"metric_fields_numerator"`
	MetricFieldsDenominator int `json:"metric_fields_denominator"`
	ScenarioNumerator       int `json:"scenario_numerator"`
	ScenarioDenominator     int `json:"scenario_denominator"`
}

type WorkspaceObservation struct {
	BeforeDigest              string   `json:"before_digest"`
	AfterDigest               string   `json:"after_digest"`
	ChangedPaths              []string `json:"changed_paths,omitempty"`
	RepositoryWorkspaceWrites bool     `json:"repository_workspace_writes"`
	MutationAuthority         string   `json:"mutation_authority"`
	MutationAuthorityReason   string   `json:"mutation_authority_reason"`
}

type Receipt struct {
	Schema                  string               `json:"schema"`
	Toolchain               string               `json:"toolchain"`
	SourceDigest            string               `json:"source_digest"`
	CanonicalSemanticDigest string               `json:"canonical_semantic_digest"`
	SourceResolution        string               `json:"source_resolution"`
	SourceIssues            []Issue              `json:"source_issues,omitempty"`
	FamilyCardinality       map[string]int       `json:"family_cardinality"`
	SourceReconstruction    SourceReconstruction `json:"source_reconstruction"`
	WorkspaceObservation    WorkspaceObservation `json:"workspace_observation"`
	Scenarios               []ScenarioResult     `json:"scenarios"`
	Digest                  string               `json:"digest"`
}
