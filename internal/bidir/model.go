package bidir

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// ID is the stable identity of a semantic node.
type ID string

// Kind is open-ended so ontology extensions do not require this package to
// know every display kind.
type Kind string

const (
	EntityKind   Kind = "Entity"
	ActivityKind Kind = "Activity"
	AgentKind    Kind = "Agent"
)

// Predicate identifies a directed semantic relation.
type Predicate string

const (
	PredicateUsed           Predicate = "prov:used"
	PredicateWasGeneratedBy Predicate = "prov:wasGeneratedBy"
	PredicateWasDerivedFrom Predicate = "prov:wasDerivedFrom"
	PredicateInvokes        Predicate = "gooo:invokes"
	PredicateRepresents     Predicate = "gooo:represents"
	PredicateSpecialization Predicate = "prov:specializationOf"
)

// SourceSpan is the dependency-free provenance boundary used by adapters.
type SourceSpan struct {
	File        string
	Start       int
	End         int
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

// Valid reports whether the span carries any source evidence.
func (s SourceSpan) Valid() bool {
	return s.File != "" || s.Start != 0 || s.End != 0
}

// Reference names a declaration from a parser-neutral document.
type Reference struct {
	ID        ID
	Name      string
	Namespace string
	Span      SourceSpan
}

// Declaration is the parser-neutral representation of a DSL declaration.
type Declaration struct {
	Kind       Kind
	ID         ID
	Name       string
	Inputs     []Reference
	Outputs    []Reference
	Attributes map[string]string
	Span       SourceSpan
}

// Document is the generic source view consumed by Get and Put.
type Document struct {
	Package      string
	Namespace    string
	Declarations []Declaration
	Relations    []Relation
}

// Node is a semantic declaration. Display fields do not define identity.
type Node struct {
	ID         ID
	Kind       Kind
	Name       string
	Namespace  string
	Aliases    []string
	Attributes map[string]string
	Span       SourceSpan
}

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
	Package    string
	Namespace  string
	Nodes      []Node
	Relations  []Relation
	Candidates FactSet
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
	n.Attributes = cloneStringMap(n.Attributes)
	return n
}

func (r Relation) normalized() Relation {
	r.ID = StableRelationID(r.Kind, r.Source, r.Target)
	r.Attributes = cloneStringMap(r.Attributes)
	return r
}

// Clone returns a detached model suitable for transactional updates.
func (m Model) Clone() Model {
	clone := Model{Package: m.Package, Namespace: m.Namespace}
	for _, node := range m.Nodes {
		clone.Nodes = append(clone.Nodes, node.normalized())
	}
	for _, relation := range m.Relations {
		clone.Relations = append(clone.Relations, relation.normalized())
	}
	clone.Candidates = m.Candidates.Normalized()
	return clone
}

// Normalized returns a deterministic, detached copy.
func (m Model) Normalized() Model {
	m = m.Clone()
	sort.Slice(m.Nodes, func(i, j int) bool {
		if m.Nodes[i].ID != m.Nodes[j].ID {
			return m.Nodes[i].ID < m.Nodes[j].ID
		}
		return m.Nodes[i].Kind < m.Nodes[j].Kind
	})
	sort.Slice(m.Relations, func(i, j int) bool {
		return relationLess(m.Relations[i], m.Relations[j])
	})
	return m
}

// Normalize canonicalizes a model in place.
func (m *Model) Normalize() {
	if m != nil {
		*m = m.Normalized()
	}
}

// Validate checks identity uniqueness and graph references.
func (m Model) Validate() error {
	seenNodes := make(map[ID]Kind, len(m.Nodes))
	for _, node := range m.Nodes {
		if err := validateID(node.ID); err != nil {
			return fmt.Errorf("node %q: %w", node.ID, err)
		}
		if node.Kind == "" {
			return fmt.Errorf("node %q has empty kind", node.ID)
		}
		if previous, exists := seenNodes[node.ID]; exists {
			return fmt.Errorf("duplicate node ID %q (%s and %s)", node.ID, previous, node.Kind)
		}
		seenNodes[node.ID] = node.Kind
	}
	seenRelations := make(map[string]struct{}, len(m.Relations))
	for _, relation := range m.Relations {
		if relation.Kind == "" {
			return fmt.Errorf("relation %q -> %q has empty predicate", relation.Source, relation.Target)
		}
		if _, exists := seenNodes[relation.Source]; !exists {
			return fmt.Errorf("relation %s references unknown source %q", relation.Kind, relation.Source)
		}
		if _, exists := seenNodes[relation.Target]; !exists {
			return fmt.Errorf("relation %s references unknown target %q", relation.Kind, relation.Target)
		}
		key := relationKey(relation.Kind, relation.Source, relation.Target)
		if _, exists := seenRelations[key]; exists {
			return fmt.Errorf("duplicate relation %s", key)
		}
		seenRelations[key] = struct{}{}
	}
	return nil
}

func validateID(id ID) error {
	if strings.TrimSpace(string(id)) == "" {
		return fmt.Errorf("empty semantic ID")
	}
	for _, r := range string(id) {
		if unicode.IsSpace(r) {
			return fmt.Errorf("semantic ID contains whitespace")
		}
	}
	return nil
}

func relationKey(predicate Predicate, source, target ID) string {
	return string(predicate) + "\x00" + string(source) + "\x00" + string(target)
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func stringMapEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func defaultName(id ID) string {
	value := string(id)
	if slash := strings.LastIndex(value, "/"); slash >= 0 && slash+1 < len(value) {
		return value[slash+1:]
	}
	return value
}
