package proofchoicealgebra

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type semanticEntry struct {
	NodeID   string `json:"node_id"`
	NodeKind string `json:"node_kind"`
	Value    Value  `json:"value"`
}

func collectValues(ir semantic.IR) (lowered, error) {
	result := lowered{Bindings: map[string]string{}}
	entries := []semanticEntry{}
	for _, node := range ir.Graph.Nodes() {
		if node.ValueProgram == "" {
			continue
		}
		result.ReconstructionDenom++
		value, err := decodeValue(node.ValueProgram)
		if err != nil {
			return result, err
		}
		result.Reconstructed++
		result.Values = append(result.Values, value)
		result.Bindings[value.ID] = node.ID.String()
		entries = append(entries, semanticEntry{node.ID.String(), node.Kind.String(), value})
	}
	if len(entries) == 0 {
		return result, fmt.Errorf("SEMANTIC_VALUE_MISSING")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].NodeID < entries[j].NodeID })
	sort.Slice(result.Values, func(i, j int) bool { return result.Values[i].ID < result.Values[j].ID })
	var b strings.Builder
	for _, entry := range entries {
		data, err := json.Marshal(canonicalEntry(entry))
		if err != nil {
			return result, fmt.Errorf("SEMANTIC_CANONICAL_UNKNOWN")
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	result.Canonical = b.String()
	result.SemanticDigest = digestBytes([]byte(result.Canonical))
	return result, nil
}

func canonicalEntry(entry semanticEntry) semanticEntry {
	entry.Value.EvidenceCapabilities = sortedRoutes(entry.Value.EvidenceCapabilities)
	entry.Value.Members = sortedCopy(entry.Value.Members)
	return entry
}

func sortedCopy(values []string) []string {
	copyOf := append([]string(nil), values...)
	sort.Strings(copyOf)
	return copyOf
}

func sortedRoutes(values []Route) []Route {
	copyOf := append([]Route(nil), values...)
	sort.Slice(copyOf, func(i, j int) bool { return copyOf[i] < copyOf[j] })
	return copyOf
}
