package languagedebugexperiment

type RuntimeObservation struct {
	Run                  int      `json:"run"`
	RuntimeReceiptSchema string   `json:"runtime_receipt_schema"`
	Runner               string   `json:"runner"`
	Toolchain            string   `json:"toolchain"`
	SourceRawDigest      string   `json:"source_raw_digest"`
	SourceSemanticDigest string   `json:"source_semantic_digest"`
	BinaryDigest         string   `json:"binary_digest"`
	Arguments            []string `json:"arguments"`
	SubjectSHA           string   `json:"subject_sha"`
	OutputDigest         string   `json:"output_digest"`
	WallNS               int64    `json:"wall_ns"`
	WallMS               int64    `json:"wall_ms"`
	PeakRSSKiB           int64    `json:"peak_rss_kib"`
}

type Measurement struct {
	Name string `json:"name"`; Executed bool `json:"executed"`; WallNS int64 `json:"wall_ns"`; WallMS int64 `json:"wall_ms"`; PeakRSSKiB int64 `json:"peak_rss_kib"`; CacheState string `json:"cache_state"`
}

type GraphEdge struct { Relation string `json:"relation"`; Subject string `json:"subject"`; Object string `json:"object"` }

type GraphObservation struct {
	Schema string `json:"schema"`; ProgramDigest string `json:"program_digest"`; GraphHash string `json:"graph_hash"`
	ActivityCount int `json:"activity_count"`; EdgeCount int `json:"edge_count"`; DebugActivityCount int `json:"debug_activity_count"`; DebugOutputCount int `json:"debug_output_count"`
	DebugUsedEdgeCount int `json:"debug_used_edge_count"`; DebugGeneratedEdgeCount int `json:"debug_generated_edge_count"`; DebugActivityIDs []string `json:"debug_activity_ids"`; DebugCausalEdges []GraphEdge `json:"debug_causal_edges"`
}
