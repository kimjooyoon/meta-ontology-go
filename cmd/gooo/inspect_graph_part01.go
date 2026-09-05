package main

type graphDump struct {
	SchemaVersion string              `json:"schema_version"`
	GraphHash     string              `json:"graph_hash"`
	SourceDigest  string              `json:"source_digest"`
	IR            graphIRStatus       `json:"ir"`
	Evidence      graphReferenceState `json:"evidence"`
	Provenance    graphReferenceState `json:"provenance"`
	Projection    graphStatus         `json:"projection"`
	Lowering      graphStatus         `json:"lowering"`
	Output        graphStatus         `json:"output"`
	Authorities   graphAuthorities    `json:"authorities"`
	Nodes         []graphNode         `json:"nodes"`
	Relations     []graphRelation     `json:"relations"`
	RuntimeBindings []graphRuntimeBinding `json:"runtime_bindings,omitempty"`
}
type graphIRStatus struct {
	Status         string `json:"status"`
	SemanticDigest string `json:"semantic_digest,omitempty"`
	Reason         string `json:"reason,omitempty"`
}
type graphReferenceState struct {
	Status string   `json:"status"`
	Refs   []string `json:"refs,omitempty"`
	Reason string   `json:"reason,omitempty"`
}
type graphStatus struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}
type graphAuthorities struct {
	GoooSource  string `json:".gooo"`
	SemanticIR  string `json:"ir"`
	Handwritten string `json:"handwritten_go"`
	Provenance  string `json:"provenance"`
	Graph       string `json:"graph"`
}
type graphNode struct {
	ID        string       `json:"id"`
	Kind      string       `json:"kind"`
	Namespace string       `json:"namespace"`
	Name      string       `json:"name"`
	Aliases   []string     `json:"aliases,omitempty"`
	Fields    []graphField `json:"fields,omitempty"`
}
type graphField struct {
	ID          string    `json:"id"`
	Parent      string    `json:"parent"`
	Name        string    `json:"name"`
	Aliases     []string  `json:"aliases,omitempty"`
	TypeRefID   string    `json:"type_ref_id"`
	Presence    string    `json:"presence"`
	Cardinality string    `json:"cardinality"`
	Source      graphSpan `json:"source"`
}
type graphRuntimeBinding struct {
	Schema           string    `json:"schema"`
	ProducerActivity string    `json:"producer_activity"`
	ProducerPort     string    `json:"producer_port"`
	ConsumerActivity string    `json:"consumer_activity"`
	ConsumerPort     string    `json:"consumer_port"`
	Entity           string    `json:"entity"`
	Source           graphSpan `json:"source"`
}
type graphSpan struct {
	File  string        `json:"file"`
	Start graphPosition `json:"start"`
	End   graphPosition `json:"end"`
}
type graphPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}
