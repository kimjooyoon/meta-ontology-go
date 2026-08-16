package freshness

import (
	"fmt"
	"sort"
)

// Kind identifies the materialized record being checked.
type Kind string

const (
	KindSource     Kind = "source"
	KindProjection Kind = "generated-projection"
	KindCache      Kind = "cache"
	KindEvidence   Kind = "evidence"
)

// State is the deterministic outcome for one record.
type State string

const (
	StateFresh   State = "fresh"
	StateStale   State = "stale"
	StateMissing State = "missing"
	StateInvalid State = "invalid"
)

// Source is an authoritative input. Digest is either measured from Path or
// supplied by an upstream authority when Path is not available.
type Source struct {
	ID     string `json:"id"`
	Path   string `json:"path,omitempty"`
	Digest string `json:"digest,omitempty"`
}

// Provenance records the PROV-shaped activity/entity boundary for a material.
// UsedIDs are references only; InputIDs on records define the freshness hash.
type Provenance struct {
	ActivityID string   `json:"activity_id"`
	EntityID   string   `json:"entity_id"`
	UsedIDs    []string `json:"used_ids,omitempty"`
}

// Artifact is a generated projection or cache entry. ContentDigest is the
// digest recorded when the artifact was produced; InputDigest is the digest
// of the ordered-by-ID current inputs at that time.
type Artifact struct {
	ID            string     `json:"id"`
	Kind          Kind       `json:"kind"`
	Path          string     `json:"path,omitempty"`
	InputIDs      []string   `json:"input_ids,omitempty"`
	InputDigest   string     `json:"input_digest"`
	ContentDigest string     `json:"content_digest"`
	Provenance    Provenance `json:"provenance"`
	EvidenceIDs   []string   `json:"evidence_ids,omitempty"`
}

// Evidence is a durable PROV Entity produced by a verification Activity. It
// uses the same digest contract as Artifact and can point at the records it
// was derived from through Provenance.UsedIDs.
type Evidence struct {
	ID            string     `json:"id"`
	Path          string     `json:"path,omitempty"`
	InputIDs      []string   `json:"input_ids,omitempty"`
	InputDigest   string     `json:"input_digest"`
	ContentDigest string     `json:"content_digest"`
	Provenance    Provenance `json:"provenance"`
}

// Requirement declares a material that must be present in a snapshot.
type Requirement struct {
	ID   string `json:"id"`
	Kind Kind   `json:"kind"`
}

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

func sortItems(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		if items[i].ID != items[j].ID {
			return items[i].ID < items[j].ID
		}
		if items[i].Path != items[j].Path {
			return items[i].Path < items[j].Path
		}
		if items[i].State != items[j].State {
			return items[i].State < items[j].State
		}
		return items[i].Detail < items[j].Detail
	})
}

func join(values []string, separator string) string {
	if len(values) == 0 {
		return ""
	}
	result := values[0]
	for _, value := range values[1:] {
		result += separator + value
	}
	return result
}
