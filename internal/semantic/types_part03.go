package semantic

import (
	"errors"
	"fmt"
	"strings"
)

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
