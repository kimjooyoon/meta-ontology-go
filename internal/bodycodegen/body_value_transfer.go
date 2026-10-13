package bodycodegen

// BodyValueCall keeps each argument occurrence distinct, including constants.
// Callee names are unresolved here; the workspace's selected typed plan owns
// activity resolution. Empty names retain unsupported call forms.
type BodyValueCall struct {
	NodeID    string   `json:"node_id"`
	Callee    string   `json:"callee"`
	Arguments []string `json:"arguments"`
}

type BodyValueTransfers struct {
	Graph BodyValueOrigins `json:"graph"`
	Calls []BodyValueCall  `json:"calls,omitempty"`
}

// InspectBodyValueTransfers reads a body without executing it. Existing local
// origin receipts keep their original shape; argument nodes belong to this view.
func InspectBodyValueTransfers(body string, arity int) BodyValueTransfers {
	graph, calls := inspectBodyValueOrigins(body, arity, true)
	return BodyValueTransfers{Graph: graph, Calls: calls}
}
