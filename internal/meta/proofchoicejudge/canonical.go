package proofchoicejudge

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type entry struct {
	NodeID   string `json:"node_id"`
	NodeKind string `json:"node_kind"`
	Value    value  `json:"value"`
}

func collectValues(ir semantic.IR) (lowered, error) {
	result := lowered{Bindings: map[string]string{}}
	entries := []entry{}
	for _, node := range ir.Graph.Nodes() {
		if node.ValueProgram == "" {
			continue
		}
		result.ReconstructionDenom++
		parsed, err := decodeValue(node.ValueProgram)
		if err != nil {
			return result, err
		}
		result.Reconstructed++
		result.Values = append(result.Values, parsed)
		result.Bindings[parsed.ID] = node.ID.String()
		entries = append(entries, entry{node.ID.String(), node.Kind.String(), parsed})
	}
	if len(entries) == 0 {
		return result, fmt.Errorf("SEMANTIC_VALUE_MISSING")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].NodeID < entries[j].NodeID })
	sort.Slice(result.Values, func(i, j int) bool { return result.Values[i].ID < result.Values[j].ID })
	var canonical strings.Builder
	for _, item := range entries {
		data, err := json.Marshal(canonicalEntry(item))
		if err != nil {
			return result, fmt.Errorf("SEMANTIC_CANONICAL_UNKNOWN")
		}
		canonical.Write(data)
		canonical.WriteByte('\n')
	}
	result.Canonical = canonical.String()
	result.SemanticDigest = digestBytes([]byte(result.Canonical))
	return result, nil
}

func canonicalEntry(item entry) entry {
	item.Value.EvidenceCapabilities = sorted(item.Value.EvidenceCapabilities)
	item.Value.Members = sorted(item.Value.Members)
	return item
}

func sorted(values []string) []string {
	copyOf := append([]string(nil), values...)
	sort.Strings(copyOf)
	return copyOf
}
