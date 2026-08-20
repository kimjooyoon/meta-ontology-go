package workfrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

func deriveR4Condensation(components []r4Component, edges []r4Edge) [][2]string {
	componentByMember := make(map[string]string)
	for _, component := range components {
		for _, member := range component.Members {
			componentByMember[member] = component.Digest
		}
	}
	seen := make(map[[2]string]struct{})
	for _, edge := range edges {
		from, to := componentByMember[edge.From], componentByMember[edge.To]
		if from == to {
			continue
		}
		seen[[2]string{from, to}] = struct{}{}
	}
	result := make([][2]string, 0, len(seen))
	for edge := range seen {
		result = append(result, edge)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i][0] != result[j][0] {
			return result[i][0] < result[j][0]
		}
		return result[i][1] < result[j][1]
	})
	return result
}
func normalizeR4Rules(rules []R4Rule) []R4Rule {
	if rules == nil {
		return nil
	}
	result := make([]R4Rule, len(rules))
	copy(result, rules)
	sort.Slice(result, func(i, j int) bool {
		if result[i].SCCDigest != result[j].SCCDigest {
			return result[i].SCCDigest < result[j].SCCDigest
		}
		if result[i].MaxIterations != result[j].MaxIterations {
			return result[i].MaxIterations < result[j].MaxIterations
		}
		return result[i].IterationsUsed < result[j].IterationsUsed
	})
	return result
}
func digestR4(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("r4 canonical digest: %v", err))
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
func edgeKey(edge r4Edge) string {
	return edge.From + "\x00" + edge.To + "\x00" + edge.PathID
}
func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
