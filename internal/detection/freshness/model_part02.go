package freshness

import (
	"fmt"
)

// Snapshot is the detector input. Paths are resolved relative to Root unless
// they are absolute. Records with no Path can still be checked by digest.
type Snapshot struct {
	Root              string        `json:"root,omitempty"`
	Sources           []Source      `json:"sources,omitempty"`
	Artifacts         []Artifact    `json:"artifacts,omitempty"`
	Evidence          []Evidence    `json:"evidence,omitempty"`
	RequiredArtifacts []Requirement `json:"required_artifacts,omitempty"`
	RequiredEvidence  []Requirement `json:"required_evidence,omitempty"`
}

// Item is one deterministic check result. Items include fresh records so a
// caller can serialize a complete status table without reconstructing it.
type Item struct {
	Kind   Kind   `json:"kind"`
	ID     string `json:"id"`
	State  State  `json:"state"`
	Path   string `json:"path,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// Report contains all results in stable kind/ID order.
type Report struct {
	Items []Item `json:"items"`
}

// Fresh reports whether every checked record is fresh.
func (r Report) Fresh() bool {
	for _, item := range r.Items {
		if item.State != StateFresh {
			return false
		}
	}
	return true
}

// Problems returns a stable copy containing only non-fresh results.
func (r Report) Problems() []Item {
	items := make([]Item, 0)
	for _, item := range r.Items {
		if item.State != StateFresh {
			items = append(items, item)
		}
	}
	return items
}

// Error returns nil for a fresh report and a deterministic aggregate error
// otherwise.
func (r Report) Error() error {
	problems := r.Problems()
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("freshness check failed: %s", formatItems(problems))
}
func formatItems(items []Item) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = string(item.Kind) + "/" + item.ID + "=" + string(item.State)
		if item.Detail != "" {
			parts[i] += " (" + item.Detail + ")"
		}
	}
	return join(parts, "; ")
}
