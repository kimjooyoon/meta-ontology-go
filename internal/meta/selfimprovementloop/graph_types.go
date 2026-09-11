package selfimprovementloop

type Graph struct {
	SchemaVersion string      `json:"schema_version"`
	GraphHash     string      `json:"graph_hash"`
	SourceDigest  string      `json:"source_digest"`
	Nodes         []GraphNode `json:"nodes"`
}

type GraphNode struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}
