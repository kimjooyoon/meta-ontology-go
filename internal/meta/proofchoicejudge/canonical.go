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
	result := lowered{}
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
		entries = append(entries, entry{node.ID.String(), node.Kind.String(), parsed})
	}
	if len(entries) == 0 {
		return result, fmt.Errorf("SEMANTIC_VALUE_MISSING")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].NodeID < entries[j].NodeID })
	var b strings.Builder
	for _, item := range entries {
		data, err := json.Marshal(canonicalEntry(item))
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

func canonicalEntry(item entry) entry {
	item.Value.Dependencies = sorted(item.Value.Dependencies)
	item.Value.Observations = sorted(item.Value.Observations)
	item.Value.AdmissibleRoutes = sorted(item.Value.AdmissibleRoutes)
	item.Value.Provenance = sorted(item.Value.Provenance)
	for index := range item.Value.Slots {
		item.Value.Slots[index].Provenance = sorted(item.Value.Slots[index].Provenance)
	}
	sort.Slice(item.Value.Slots, func(i, j int) bool { return item.Value.Slots[i].ID < item.Value.Slots[j].ID })
	return item
}

func sorted(values []string) []string {
	copyOf := append([]string(nil), values...)
	sort.Strings(copyOf)
	return copyOf
}
