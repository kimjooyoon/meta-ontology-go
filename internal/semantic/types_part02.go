package semantic

// Node is a semantic declaration. Name and Aliases are presentation and
// lookup metadata; ID, Kind, Namespace, and latent field structure are the
// semantic identity boundary. Fields are valid only on Entity nodes.
type Node struct {
	ID        ID
	Kind      Kind
	Namespace Namespace
	Name      string
	Aliases   []string
	Fields    []Field `json:"fields,omitempty"`
	Span      Span
}

func NewNode(kind Kind, id ID, namespace Namespace, name string) (Node, error) {
	node := Node{ID: id, Kind: kind, Namespace: namespace, Name: name}
	return node.Normalized()
}
func NewNodeFromStrings(kind Kind, id, namespace, name string) (Node, error) {
	parsedID, err := ParseIdentity(id)
	if err != nil {
		return Node{}, err
	}
	parsedNamespace, err := ParseNamespace(namespace)
	if err != nil {
		return Node{}, err
	}
	return NewNode(kind, parsedID, parsedNamespace, name)
}
func NewEntity(id ID, namespace Namespace, name string) (Node, error) {
	return NewNode(Entity, id, namespace, name)
}
func NewActivity(id ID, namespace Namespace, name string) (Node, error) {
	return NewNode(Activity, id, namespace, name)
}
func NewAgent(id ID, namespace Namespace, name string) (Node, error) {
	return NewNode(Agent, id, namespace, name)
}
