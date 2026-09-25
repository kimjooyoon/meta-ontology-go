package bidir

import (
	"sort"
	"strings"
)

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
