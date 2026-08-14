package semantic

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrInvalidNode = errors.New("invalid semantic node")
	ErrInvalidSpan = errors.New("invalid source span")
)

// Kind is the PROV-inspired role of a semantic node.
type Kind string

const (
	Entity   Kind = "Entity"
	Activity Kind = "Activity"
	Agent    Kind = "Agent"

	EntityKind   = Entity
	ActivityKind = Activity
	AgentKind    = Agent
)

func (k Kind) String() string {
	return string(k)
}

func (k Kind) Valid() bool {
	switch k {
	case Entity, Activity, Agent:
		return true
	default:
		return false
	}
}

// Position is a source position. Zero means unknown; offsets, lines, and
// columns are otherwise expected to be non-negative.
type Position struct {
	Offset int
	Line   int
	Column int
}

func (p Position) valid() bool {
	return p.Offset >= 0 && p.Line >= 0 && p.Column >= 0
}

// Span records source provenance without coupling the semantic package to a
// lexer or parser implementation.
type Span struct {
	File  string
	Start Position
	End   Position
}

func (s Span) IsZero() bool {
	return s.File == "" && s.Start == (Position{}) && s.End == (Position{})
}

func (s Span) Normalized() Span {
	s.File = strings.TrimSpace(s.File)
	return s
}

func (s Span) Validate() error {
	if !s.Start.valid() || !s.End.valid() {
		return fmt.Errorf("%w: positions must be non-negative", ErrInvalidSpan)
	}
	if s.End.Offset < s.Start.Offset {
		return fmt.Errorf("%w: end offset precedes start offset", ErrInvalidSpan)
	}
	return nil
}

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

func (n Node) Normalized() (Node, error) {
	id, err := ParseIdentity(n.ID.String())
	if err != nil {
		return Node{}, fmt.Errorf("%w: id: %v", ErrInvalidNode, err)
	}
	ns, err := ParseNamespace(n.Namespace.String())
	if err != nil {
		return Node{}, fmt.Errorf("%w: namespace: %v", ErrInvalidNode, err)
	}
	if !n.Kind.Valid() {
		return Node{}, fmt.Errorf("%w: unknown kind %q", ErrInvalidNode, n.Kind)
	}
	name, err := normalizeName(n.Name)
	if err != nil {
		return Node{}, fmt.Errorf("%w: name: %v", ErrInvalidNode, err)
	}
	aliases, err := normalizeAliases(n.Aliases, name)
	if err != nil {
		return Node{}, fmt.Errorf("%w: aliases: %v", ErrInvalidNode, err)
	}
	span := n.Span.Normalized()
	if err := span.Validate(); err != nil {
		return Node{}, fmt.Errorf("%w: span: %v", ErrInvalidNode, err)
	}
	fields, err := normalizeFields(n.Fields, id, n.Kind)
	if err != nil {
		return Node{}, fmt.Errorf("%w: fields: %w", ErrInvalidNode, err)
	}

	n.ID = id
	n.Namespace = ns
	n.Name = name
	n.Aliases = aliases
	n.Fields = fields
	n.Span = span
	return n, nil
}

func (n Node) Validate() error {
	_, err := n.Normalized()
	return err
}

func (n Node) NameRef() NameRef {
	return NameRef{Namespace: n.Namespace, Name: n.Name}
}

func (n Node) HasName(name string) bool {
	canonical, err := normalizeName(name)
	if err != nil {
		return false
	}
	if n.Name == canonical {
		return true
	}
	for _, alias := range n.Aliases {
		if alias == canonical {
			return true
		}
	}
	return false
}

func normalizeName(raw string) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")
	if name == "" {
		return "", errors.New("name is empty")
	}
	return name, nil
}

func normalizeAliases(raw []string, name string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(raw))
	aliases := make([]string, 0, len(raw))
	for _, value := range raw {
		alias, err := normalizeName(value)
		if err != nil {
			return nil, err
		}
		if alias == name {
			continue
		}
		if _, exists := seen[alias]; exists {
			continue
		}
		seen[alias] = struct{}{}
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	return aliases, nil
}
