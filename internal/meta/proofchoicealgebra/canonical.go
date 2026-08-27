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
	result := lowered{}
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
		entries = append(entries, semanticEntry{node.ID.String(), node.Kind.String(), value})
	}
	if len(entries) == 0 {
		return result, fmt.Errorf("SEMANTIC_VALUE_MISSING")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].NodeID < entries[j].NodeID })
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
	entry.Value.Dependencies = sortedCopy(entry.Value.Dependencies)
	entry.Value.Observations = sortedCopy(entry.Value.Observations)
	entry.Value.AdmissibleRoutes = sortedRoutes(entry.Value.AdmissibleRoutes)
	entry.Value.Provenance = sortedCopy(entry.Value.Provenance)
	for index := range entry.Value.Slots {
		entry.Value.Slots[index].Provenance = sortedCopy(entry.Value.Slots[index].Provenance)
	}
	sort.Slice(entry.Value.Slots, func(i, j int) bool { return entry.Value.Slots[i].ID < entry.Value.Slots[j].ID })
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
