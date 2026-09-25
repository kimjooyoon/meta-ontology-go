package query

// Metadata returns a detached, current snapshot of the query projection. The
// graph hash follows the current view, while SemanticDigest remains bound to
// the IR snapshot from which the view was derived.
func (graph Graph) Metadata() ProjectionMetadata {
	metadata := ProjectionMetadata{
		SchemaVersion:    QueryProjectionSchemaVersion,
		Namespace:        "",
		GraphHash:        graph.StableHash(),
		SourceStatus:     "unavailable",
		EvidenceStatus:   "unknown",
		ProvenanceStatus: "unknown",
		ProjectionStatus: "unbound",
		DerivedStatus:    DerivedStatusNotRequested,
		AuthorityLabels: []AuthorityLabel{
			{View: ".gooo", Authority: "authoritative", Status: "unavailable"},
			{View: "semantic_ir", Authority: "authoritative", Status: "unavailable"},
			{View: "handwritten_go", Authority: "authoritative", Status: "unavailable"},
			{View: "generated_go", Authority: "derived", Status: "unavailable"},
			{View: "provenance", Authority: "authoritative", Status: "unknown"},
			{View: "query_graph", Authority: "derived", Status: "unbound"},
			{View: "derived_query", Authority: "derived", Status: DerivedStatusNotRequested},
		},
	}
	if graph.binding == nil {
		return metadata
	}
	metadata.SemanticDigest = graph.binding.semanticDigest
	metadata.Namespace = graph.binding.namespace
	metadata.SourceDigest = graph.binding.sourceDigest
	metadata.EvidenceDigest = graph.binding.evidenceDigest
	metadata.ProvenanceDigest = graph.binding.provenanceDigest
	metadata.SourceStatus = graph.binding.sourceStatus
	metadata.EvidenceStatus = graph.binding.evidenceStatus
	metadata.ProvenanceStatus = graph.binding.provenanceStatus
	metadata.ProjectionStatus = "derived"
	for index := range metadata.AuthorityLabels {
		switch metadata.AuthorityLabels[index].View {
		case "semantic_ir":
			metadata.AuthorityLabels[index].Status = "bound"
		case "provenance":
			metadata.AuthorityLabels[index].Status = metadata.ProvenanceStatus
		case "query_graph":
			metadata.AuthorityLabels[index].Status = "current"
		}
	}
	return metadata
}
