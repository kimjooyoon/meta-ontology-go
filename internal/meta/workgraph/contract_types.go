package workgraph

type GateSpec struct {
	ID          string `json:"id"`
	Activity    string `json:"activity"`
	Stage       string `json:"stage"`
	Step        string `json:"step"`
	EvidenceKey string `json:"evidence_key"`
	ProofChoice string `json:"proof_choice"`
}

type ClaimSpec struct {
	ID        string `json:"id"`
	Entity    string `json:"entity"`
	Authority string `json:"authority"`
}

type Contract struct {
	Schema  string     `json:"schema"`
	Project string     `json:"project"`
	Source  string     `json:"source"`
	Claim   ClaimSpec  `json:"claim"`
	Gates   []GateSpec `json:"gates"`
}

type ResourceSample struct {
	Samples          int    `json:"samples"`
	WallNanoseconds  int64  `json:"wall_nanoseconds"`
	HeapSysBytes     uint64 `json:"heap_sys_bytes"`
	TotalAllocBytes  uint64 `json:"total_alloc_bytes"`
}
