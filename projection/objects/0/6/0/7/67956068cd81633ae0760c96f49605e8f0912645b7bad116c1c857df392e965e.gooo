package couplingexplain

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Control carries both requested and observed values. Any race is UNKNOWN.
type Control struct {
	RequestVersion              uint64 `json:"request_version"`
	ObservedVersion             uint64 `json:"observed_version"`
	RequestCancellationVersion  uint64 `json:"request_cancellation_version"`
	ObservedCancellationVersion uint64 `json:"observed_cancellation_version"`
}

// SnapshotBinding is stable input identity; it contains no presentation data.
type SnapshotBinding struct {
	SnapshotDigest       string  `json:"snapshot_digest"`
	RegistryDigest       string  `json:"registry_digest"`
	SourceMapDigest      string  `json:"source_map_digest"`
	ManifestDigest       string  `json:"manifest_digest"`
	ToolchainDigest      string  `json:"toolchain_digest"`
	ProfileDigest        string  `json:"profile_digest"`
	DetectorInputDigest  string  `json:"detector_input_digest"`
	DetectorResultDigest string  `json:"detector_result_digest"`
	VerifierResultDigest string  `json:"verifier_result_digest"`
	EnvelopeDigest       string  `json:"envelope_digest"`
	Control              Control `json:"control"`
}

// Presentation is accepted at the adapter boundary but never enters an
// evidence digest or canonical explanation output.
type Presentation struct {
	Label     string `json:"label,omitempty"`
	Root      string `json:"root,omitempty"`
	Path      string `json:"path,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Actor     string `json:"actor,omitempty"`
}
type CodeBindingSummary struct {
	CodeSymbolID        string       `json:"code_symbol_id"`
	SemanticOwnerID     string       `json:"semantic_owner_id"`
	RegisteredSurfaceID string       `json:"registered_surface_id"`
	SourceMapID         string       `json:"source_map_id"`
	BindingDigest       string       `json:"binding_digest"`
	CodeBindingDigest   string       `json:"code_binding_digest"`
	Presentation        Presentation `json:"presentation"`
}
type TermSummary struct {
	TermID           string       `json:"term_id"`
	SemanticOwnerID  string       `json:"semantic_owner_id"`
	Version          string       `json:"version"`
	DefinitionDigest string       `json:"definition_digest"`
	Presentation     Presentation `json:"presentation"`
}
type PathStep struct {
	FromID       string                  `json:"from_id"`
	ToID         string                  `json:"to_id"`
	Kind         semantic.InferenceKind  `json:"kind"`
	Phase        semantic.PhasePlacement `json:"phase"`
	RuleRef      string                  `json:"rule_ref,omitempty"`
	InputDigest  string                  `json:"input_digest"`
	OutputDigest string                  `json:"output_digest"`
	EvidenceRef  string                  `json:"evidence_ref,omitempty"`
}
type PathSummary struct {
	PathID       string       `json:"path_id"`
	StartID      string       `json:"start_id"`
	EndID        string       `json:"end_id"`
	StepCount    int          `json:"step_count"`
	PathDigest   string       `json:"path_digest"`
	Steps        []PathStep   `json:"steps,omitempty"`
	Presentation Presentation `json:"presentation"`
}
