package operationprovenance

type Node struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type Metric struct {
	ID        string   `json:"id"`
	Family    string   `json:"family"`
	Claim     string   `json:"claim"`
	DependsOn []string `json:"depends_on,omitempty"`
}

type Fixture struct {
	ID      string   `json:"id"`
	Nodes   []Node   `json:"nodes"`
	Edges   []Edge   `json:"edges"`
	Metrics []Metric `json:"metrics"`
}

type Lineage struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	EvidencePath  string `json:"evidence_path"`
}

type Issue struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Cause     string   `json:"cause"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}

type MetricResult struct {
	ID              string  `json:"id"`
	Family          string  `json:"family"`
	Claim           string  `json:"claim"`
	Numerator       int     `json:"numerator"`
	Denominator     int     `json:"denominator"`
	Decision        string  `json:"decision"`
	EvaluationState string  `json:"evaluation_state"`
	Lineage         Lineage `json:"lineage"`
	Issue           *Issue  `json:"issue,omitempty"`
}

type GraphSummary struct {
	Nodes     int            `json:"nodes"`
	Edges     int            `json:"edges"`
	EdgeKinds map[string]int `json:"edge_kinds"`
}

type ScenarioResult struct {
	ID          string         `json:"id"`
	Graph       GraphSummary   `json:"graph"`
	Numerator   int            `json:"numerator"`
	Denominator int            `json:"denominator"`
	Decisions   map[string]int `json:"decisions"`
	Metrics     []MetricResult `json:"metrics"`
}

type Receipt struct {
	Schema                    string           `json:"schema"`
	Toolchain                 string           `json:"toolchain"`
	RepositoryWorkspaceWrites bool             `json:"repository_workspace_writes"`
	MutationAuthority         bool             `json:"mutation_authority"`
	SourceDigest              string           `json:"source_digest"`
	Scenarios                 []ScenarioResult `json:"scenarios"`
	Digest                    string           `json:"digest"`
}

type Report struct {
	Schema           string `json:"schema"`
	Status           string `json:"status"`
	SourceDigest     string `json:"source_digest"`
	ReceiptDigest    string `json:"receipt_digest"`
	ScenarioCount    int    `json:"scenario_count"`
	MetricCount      int    `json:"metric_count"`
	FailClosedCount  int    `json:"fail_closed_count"`
	DirectUnknowns   int    `json:"direct_unknowns"`
	DependencyBlocks int    `json:"dependency_blocks"`
	Digest           string `json:"digest"`
}
