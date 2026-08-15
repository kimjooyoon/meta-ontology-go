package bidir

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func nodeSemanticEqual(left, right Node) bool {
	if left.ID != right.ID || left.Kind != right.Kind || !stringMapEqual(left.Attributes, right.Attributes) || len(left.Fields) != len(right.Fields) {
		return false
	}
	for index := range left.Fields {
		if !fieldSemanticEqual(left.Fields[index], right.Fields[index]) {
			return false
		}
	}
	return true
}

func relationSemanticEqual(left, right Relation) bool {
	return left.Kind == right.Kind && left.Source == right.Source && left.Target == right.Target && stringMapEqual(left.Attributes, right.Attributes)
}

// SemanticEquivalent ignores presentation, source spans, and candidates.
func SemanticEquivalent(left, right Model) bool {
	left, right = left.Normalized(), right.Normalized()
	if len(left.Nodes) != len(right.Nodes) || len(left.Relations) != len(right.Relations) {
		return false
	}
	for index := range left.Nodes {
		if !nodeSemanticEqual(left.Nodes[index], right.Nodes[index]) {
			return false
		}
	}
	for index := range left.Relations {
		if !relationSemanticEqual(left.Relations[index], right.Relations[index]) {
			return false
		}
	}
	return true
}

// SemanticallyEquivalent is the method form of SemanticEquivalent.
func (m Model) SemanticallyEquivalent(other Model) bool {
	return SemanticEquivalent(m, other)
}

// SemanticFingerprint is a stable digest of the fields used by equivalence.
func SemanticFingerprint(model Model) string {
	model = model.Normalized()
	var canonical strings.Builder
	for _, node := range model.Nodes {
		writeFingerprintPart(&canonical, string(node.ID))
		writeFingerprintPart(&canonical, string(node.Kind))
		writeMapFingerprint(&canonical, node.Attributes)
		for _, field := range node.Fields {
			writeFieldSemanticFingerprint(&canonical, field)
		}
	}
	canonical.WriteByte('|')
	for _, relation := range model.Relations {
		writeFingerprintPart(&canonical, string(relation.Kind))
		writeFingerprintPart(&canonical, string(relation.Source))
		writeFingerprintPart(&canonical, string(relation.Target))
		writeMapFingerprint(&canonical, relation.Attributes)
	}
	digest := sha256.Sum256([]byte(canonical.String()))
	return hex.EncodeToString(digest[:])
}

func writeFingerprintPart(builder *strings.Builder, value string) {
	fmt.Fprintf(builder, "%d:", len(value))
	builder.WriteString(value)
}

func writeMapFingerprint(builder *strings.Builder, values map[string]string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteByte('{')
		writeFingerprintPart(builder, key)
		writeFingerprintPart(builder, values[key])
		builder.WriteByte('}')
	}
}

// Diff computes a presentation-insensitive semantic delta.
func Diff(before, after Model) Delta {
	before, after = before.Normalized(), after.Normalized()
	delta := Delta{}
	deltaNodes(&delta, before.Nodes, after.Nodes)
	deltaRelations(&delta, before.Relations, after.Relations)
	delta.Normalize()
	return delta
}

func deltaNodes(delta *Delta, before, after []Node) {
	left, right := nodeMap(before), nodeMap(after)
	for id, node := range left {
		other, exists := right[id]
		if !exists || !nodeSemanticEqual(node, other) {
			delta.RemovedNodes = append(delta.RemovedNodes, node)
		}
	}
	for id, node := range right {
		other, exists := left[id]
		if !exists || !nodeSemanticEqual(node, other) {
			delta.AddedNodes = append(delta.AddedNodes, node)
		}
	}
}

func deltaRelations(delta *Delta, before, after []Relation) {
	left, right := relationMap(before), relationMap(after)
	for key, relation := range left {
		other, exists := right[key]
		if !exists || !relationSemanticEqual(relation, other) {
			delta.RemovedRelations = append(delta.RemovedRelations, relation)
		}
	}
	for key, relation := range right {
		other, exists := left[key]
		if !exists || !relationSemanticEqual(relation, other) {
			delta.AddedRelations = append(delta.AddedRelations, relation)
		}
	}
}

func nodeMap(nodes []Node) map[ID]Node {
	result := make(map[ID]Node, len(nodes))
	for _, node := range nodes {
		result[node.ID] = node
	}
	return result
}

func relationMap(relations []Relation) map[string]Relation {
	result := make(map[string]Relation, len(relations))
	for _, relation := range relations {
		result[relationKey(relation.Kind, relation.Source, relation.Target)] = relation
	}
	return result
}

// Normalize sorts delta members and recalculates relation IDs.
func (d *Delta) Normalize() {
	if d == nil {
		return
	}
	for index := range d.AddedNodes {
		d.AddedNodes[index] = d.AddedNodes[index].normalized()
	}
	for index := range d.RemovedNodes {
		d.RemovedNodes[index] = d.RemovedNodes[index].normalized()
	}
	for index := range d.AddedRelations {
		d.AddedRelations[index] = d.AddedRelations[index].normalized()
	}
	for index := range d.RemovedRelations {
		d.RemovedRelations[index] = d.RemovedRelations[index].normalized()
	}
	sort.Slice(d.AddedNodes, func(i, j int) bool { return d.AddedNodes[i].ID < d.AddedNodes[j].ID })
	sort.Slice(d.RemovedNodes, func(i, j int) bool { return d.RemovedNodes[i].ID < d.RemovedNodes[j].ID })
	sort.Slice(d.AddedRelations, func(i, j int) bool { return relationLess(d.AddedRelations[i], d.AddedRelations[j]) })
	sort.Slice(d.RemovedRelations, func(i, j int) bool { return relationLess(d.RemovedRelations[i], d.RemovedRelations[j]) })
}

// IsEmpty reports whether the delta changes semantic content.
func (d Delta) IsEmpty() bool {
	return len(d.AddedNodes) == 0 && len(d.RemovedNodes) == 0 && len(d.AddedRelations) == 0 && len(d.RemovedRelations) == 0
}
