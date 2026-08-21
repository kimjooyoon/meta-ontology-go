package bidir

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
