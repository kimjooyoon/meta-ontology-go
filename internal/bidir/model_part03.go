package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

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

// Validate checks identity uniqueness, typed fields, and graph references.
func (m Model) Validate() error {
	return m.ValidateWithTypes(semantic.DefaultTypeRegistry())
}
