package operationprovenance

type MetricResult struct {
	ID                string          `json:"id"`
	Family            string          `json:"family"`
	Claim             string          `json:"claim"`
	Numerator         int             `json:"numerator"`
	Denominator       int             `json:"denominator"`
	Decision          string          `json:"decision"`
	SubjectResolution string          `json:"subject_resolution"`
	EvaluationState   string          `json:"evaluation_state"`
	Lineage           Lineage         `json:"lineage"`
	Issue             *Issue          `json:"issue,omitempty"`
	Transition        ClaimTransition `json:"claim_transition"`
}

type GraphSummary struct {
	Nodes     int            `json:"nodes"`
	Edges     int            `json:"edges"`
	EdgeKinds map[string]int `json:"edge_kinds"`
}

type ScenarioResult struct {
	ID                  string         `json:"id"`
	Mutation            string         `json:"mutation"`
	Graph               GraphSummary   `json:"graph"`
	Numerator           int            `json:"numerator"`
	Denominator         int            `json:"denominator"`
	ConformanceDecision string         `json:"conformance_decision"`
	SubjectResolution   string         `json:"subject_resolution"`
	Decisions           map[string]int `json:"decisions"`
	Transitions         map[string]int `json:"transitions"`
	Metrics             []MetricResult `json:"metrics"`
}
