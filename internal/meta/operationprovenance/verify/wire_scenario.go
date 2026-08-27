package verify

type graphSummary struct {
	Nodes     int            `json:"nodes"`
	Edges     int            `json:"edges"`
	EdgeKinds map[string]int `json:"edge_kinds"`
}
type scenarioResult struct {
	ID          string         `json:"id"`
	Mutation    string         `json:"mutation"`
	Graph       graphSummary   `json:"graph"`
	Numerator   int            `json:"numerator"`
	Denominator int            `json:"denominator"`
	Decision    string         `json:"conformance_decision"`
	Resolution  string         `json:"subject_resolution"`
	Decisions   map[string]int `json:"decisions"`
	Transitions map[string]int `json:"transitions"`
	Metrics     []metricResult `json:"metrics"`
}
type sourceReconstruction struct {
	Numerator               int `json:"numerator"`
	Denominator             int `json:"denominator"`
	MetricFieldsNumerator   int `json:"metric_fields_numerator"`
	MetricFieldsDenominator int `json:"metric_fields_denominator"`
	ScenarioNumerator       int `json:"scenario_numerator"`
	ScenarioDenominator     int `json:"scenario_denominator"`
}
type workspaceObservation struct {
	Before    string   `json:"before_digest"`
	After     string   `json:"after_digest"`
	Changed   []string `json:"changed_paths,omitempty"`
	Writes    bool     `json:"repository_workspace_writes"`
	Authority bool     `json:"mutation_authority"`
}
type receipt struct {
	Schema         string               `json:"schema"`
	Toolchain      string               `json:"toolchain"`
	Source         string               `json:"source_digest"`
	Semantic       string               `json:"canonical_semantic_digest"`
	Reconstruction sourceReconstruction `json:"source_reconstruction"`
	Observation    workspaceObservation `json:"workspace_observation"`
	Scenarios      []scenarioResult     `json:"scenarios"`
	Digest         string               `json:"digest"`
}
