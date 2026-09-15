package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
)

// SnapshotInput contains all authority needed to construct a Snapshot. A nil
// RegisteredIDs slice is deliberately different from an empty registry: the
// former means registry membership was not supplied and is UNKNOWN.
type SnapshotInput struct {
	Sources         []SourceInput
	SourceMapDigest string
	RegistryDigest  string
	RegisteredIDs   []string

	// CandidateBindings and DerivedBindings are observations, not authoritative
	// semantic bindings. Supplying either is rejected rather than promoted or
	// silently ignored.
	CandidateBindings []semanticbinding.Binding
	DerivedBindings   []semanticbinding.Binding
}

// Input is the concise spelling of SnapshotInput.
type Input = SnapshotInput

// ManifestInput is the vocabulary alias for SnapshotInput.
type ManifestInput = SnapshotInput

// Binding is the explicit, source-bound manifest representation of a
// semanticbinding.Binding.
type Binding struct {
	ID            string               `json:"id"`
	Role          semanticbinding.Role `json:"role"`
	Status        Status               `json:"status"`
	BindingDigest string               `json:"binding_digest"`
}

// Source is one canonical manifest source record.
type Source struct {
	Path       string `json:"path"`
	BlobDigest string `json:"blob_digest"`
	// Bindings is [] for an explicit unbound source and never null in a
	// canonical snapshot.
	Bindings []Binding `json:"bindings"`
}

// Snapshot is a canonical, source-backed manifest. A BOUND snapshot always
// has a valid Digest that covers every field except Digest itself.
type Snapshot struct {
	Schema            string   `json:"schema"`
	Status            Status   `json:"status"`
	FullSuiteFallback bool     `json:"full_suite_fallback"`
	SourceMapDigest   string   `json:"source_map_digest"`
	RegistryDigest    string   `json:"registry_digest"`
	RegisteredIDs     []string `json:"registered_ids"`
	Sources           []Source `json:"sources"`
	Digest            string   `json:"digest"`
}

// SnapshotManifest is the persisted-manifest spelling of Snapshot.
type SnapshotManifest = Snapshot

// SourceRecord and BindingRecord are explicit manifest vocabulary aliases.
type SourceRecord = Source
type BindingRecord = Binding

// Delta is the source-backed union of stable IDs whose exact binding changed.
// ChangedIDs is empty for an exact replay and also for every UNKNOWN result.
type Delta struct {
	Status            Status   `json:"status"`
	ChangedIDs        []string `json:"changed_ids"`
	FullSuiteFallback bool     `json:"full_suite_fallback"`
}
