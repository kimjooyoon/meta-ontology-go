package pressureshadow

import (
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
	"sort"
)

func checkUniqueID(seen map[string]struct{}, id, kind string) error {
	if !validID(id) {
		return fmt.Errorf("invalid %s ID", kind)
	}
	if _, exists := seen[id]; exists {
		return fmt.Errorf("duplicate %s ID %q", kind, id)
	}
	seen[id] = struct{}{}
	return nil
}
func CanonicalInputBytes(input Input) ([]byte, error) {
	if err := validateSyntax(input); err != nil {
		return nil, err
	}
	rows := append([]PathCoverage{}, input.PathCoverage...)
	sort.Slice(rows, func(left, right int) bool { return rows[left].PathID < rows[right].PathID })
	type canonicalRow struct {
		PathID         string          `json:"path_id"`
		SnapshotDigest string          `json:"snapshot_digest"`
		PolicyDigest   string          `json:"policy_digest"`
		RegistryDigest string          `json:"registry_digest"`
		Coverage       json.RawMessage `json:"coverage"`
	}
	canonicalRows := make([]canonicalRow, 0, len(rows))
	for _, row := range rows {
		coverage, err := pressurecoverage.CanonicalInputBytes(row.Coverage)
		if err != nil {
			return nil, err
		}
		canonicalRows = append(canonicalRows, canonicalRow{
			PathID: row.PathID, SnapshotDigest: row.SnapshotDigest,
			PolicyDigest: row.PolicyDigest, RegistryDigest: row.RegistryDigest, Coverage: coverage,
		})
	}
	return json.Marshal(struct {
		Schema       string             `json:"schema"`
		Selector     workfrontier.Input `json:"selector"`
		PathCoverage []canonicalRow     `json:"path_coverage"`
	}{input.Schema, canonicalSelector(input.Selector), canonicalRows})
}
func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}
