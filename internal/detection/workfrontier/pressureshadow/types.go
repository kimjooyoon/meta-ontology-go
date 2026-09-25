// Package pressureshadow contains the S1a1 strict canonical bridge envelope.
// Zero rows, blank tuple values, and semantically mismatched sets are
// canonical data here; they are not VALID/PASS or completeness decisions.
// S1a2 owns semantic validation, and S1b owns selector/A2 observation.
package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

const SchemaVersion = "gooo/workfrontier-pressure-shadow/v1"

type PathCoverage struct {
	PathID         string                 `json:"path_id"`
	SnapshotDigest string                 `json:"snapshot_digest"`
	PolicyDigest   string                 `json:"policy_digest"`
	RegistryDigest string                 `json:"registry_digest"`
	Coverage       pressurecoverage.Input `json:"coverage"`
}

type Input struct {
	Schema       string             `json:"schema"`
	Selector     workfrontier.Input `json:"selector"`
	PathCoverage []PathCoverage     `json:"path_coverage"`
}
