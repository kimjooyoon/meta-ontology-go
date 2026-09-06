package bidir

import (
	"encoding/hex"
	"sort"
)

// Relation is a directed semantic edge.
type Relation struct {
	ID         ID
	Kind       Predicate
	Source     ID
	Target     ID
	Attributes map[string]string
	Span       SourceSpan
}

// Model is the normalized semantic view used by the generic lens.
type Model struct {
	Package         string
	Namespace       string
	Nodes           []Node
	Relations       []Relation
	RuntimeBindings []RuntimeBinding
	Candidates      FactSet
	// Activity port arities preserve source declaration cardinality that is
	// otherwise lost when equal PROV edges are normalized into one relation.
	activityInputArity  map[ID]int
	activityOutputArity map[ID]int
}

// Delta is a normalized semantic change set.
type Delta struct {
	AddedNodes       []Node
	RemovedNodes     []Node
	AddedRelations   []Relation
	RemovedRelations []Relation
}

// Locality describes a delta's touched nodes and one-hop affected neighbors.
type Locality struct {
	Touched  []ID
	Affected []ID
}

// NewEntity returns an initialized entity node.
func NewEntity(id ID, name, namespace string) Node {
	return Node{ID: id, Kind: EntityKind, Name: name, Namespace: namespace}
}

// NewActivity returns an initialized activity node.
func NewActivity(id ID, name, namespace string) Node {
	return Node{ID: id, Kind: ActivityKind, Name: name, Namespace: namespace}
}

// StableRelationID makes edge identity independent of source formatting.
func StableRelationID(predicate Predicate, source, target ID) ID {
	encode := func(value string) string {
		return hex.EncodeToString([]byte(value))
	}
	return ID("urn:gooo:relation:" + encode(string(predicate)) + ":" + encode(string(source)) + ":" + encode(string(target)))
}
func (n Node) normalized() Node {
	n.Aliases = append([]string(nil), n.Aliases...)
	sort.Strings(n.Aliases)
	n.Fields = cloneFields(n.Fields)
	n.Attributes = cloneStringMap(n.Attributes)
	return n
}
func (r Relation) normalized() Relation {
	r.ID = StableRelationID(r.Kind, r.Source, r.Target)
	r.Attributes = cloneStringMap(r.Attributes)
	return r
}
