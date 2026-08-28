package languageutility

type CellObservation struct {
	UseCaseID      string `json:"use_case_id"`
	StageID        string `json:"stage_id"`
	State          string `json:"state"`
	Producer       string `json:"producer"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceKey    string `json:"evidence_key,omitempty"`
	EvidencePath   string `json:"evidence_path,omitempty"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
	MetaActivityID       string       `json:"meta_activity_id,omitempty"`
	MetaInputID          string       `json:"meta_input_id,omitempty"`
	MetaOutputID         string       `json:"meta_output_id,omitempty"`
	ActivityMatches      int          `json:"activity_matches,omitempty"`
	OutputMatches        int          `json:"output_matches,omitempty"`
	UsedEdgeMatches      int          `json:"used_edge_matches,omitempty"`
	GeneratedEdgeMatches int          `json:"generated_edge_matches,omitempty"`
	CausalEdges          []CausalEdge `json:"causal_edges,omitempty"`
}

type CausalEdge struct {
	Relation string `json:"relation"`
	Subject  string `json:"subject"`
	Object   string `json:"object"`
}

type GraphObservation struct {
	Schema                  string `json:"schema"`
	ProgramDigest           string `json:"program_digest"`
	GraphHash               string `json:"graph_hash"`
	ActivityCount           int    `json:"activity_count"`
	EdgeCount               int    `json:"edge_count"`
	DebugActivityCount      int    `json:"debug_activity_count"`
	DebugOutputCount        int    `json:"debug_output_count"`
	DebugUsedEdgeCount      int    `json:"debug_used_edge_count"`
	DebugGeneratedEdgeCount int    `json:"debug_generated_edge_count"`
	DebugActivityIDs        []string    `json:"debug_activity_ids"`
	DebugCausalEdges        []CausalEdge `json:"debug_causal_edges"`
}

type Observation struct {
	Schema           string            `json:"schema"`
	ContractID       string            `json:"contract_id"`
	SubjectSHA       string            `json:"subject_sha"`
	RepositoryWrites int               `json:"repository_writes"`
	Graph            GraphObservation `json:"graph"`
	Cells            []CellObservation `json:"cells"`
}

const (
	StateClosed  = "CLOSED"
	StateOpen    = "OPEN"
	StateUnknown = "UNKNOWN"
	StateRefuted = "REFUTED"
)
