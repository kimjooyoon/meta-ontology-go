package freshness

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
