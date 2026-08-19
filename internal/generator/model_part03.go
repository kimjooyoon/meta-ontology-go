package generator

import (
	"encoding/json"
)

// MarshalJSON keeps the established fieldless source-map wire form stable
// while publishing the additional field metadata only for field mappings.
func (mapping SourceMapping) MarshalJSON() ([]byte, error) {
	type sourceMappingWire struct {
		SemanticID     string      `json:"SemanticID"`
		Kind           string      `json:"Kind"`
		Ordinal        int         `json:"Ordinal"`
		Source         SourceSpan  `json:"Source"`
		Generated      SourceRange `json:"Generated"`
		ParentID       string      `json:"parent_id,omitempty"`
		TypeRefID      string      `json:"type_ref_id,omitempty"`
		Presence       string      `json:"presence,omitempty"`
		Cardinality    string      `json:"cardinality,omitempty"`
		NameSource     *SourceSpan `json:"name_source,omitempty"`
		ProfileID      string      `json:"profile_id,omitempty"`
		ProfileVersion int         `json:"profile_version,omitempty"`
		ProfileDigest  string      `json:"profile_digest,omitempty"`
	}
	var nameSource *SourceSpan
	if !sourceSpanIsZero(mapping.NameSource) {
		span := mapping.NameSource
		nameSource = &span
	}
	return json.Marshal(sourceMappingWire{
		SemanticID: mapping.SemanticID, Kind: mapping.Kind, Ordinal: mapping.Ordinal,
		Source: mapping.Source, Generated: mapping.Generated, ParentID: mapping.ParentID,
		TypeRefID: mapping.TypeRefID, Presence: mapping.Presence, Cardinality: mapping.Cardinality,
		NameSource: nameSource, ProfileID: mapping.ProfileID, ProfileVersion: mapping.ProfileVersion,
		ProfileDigest: mapping.ProfileDigest,
	})
}

// Lookup returns mappings for an identity in generated order.
func (m SourceMap) Lookup(semanticID string) []SourceMapping {
	var result []SourceMapping
	for _, mapping := range m.Mappings {
		if mapping.SemanticID == semanticID {
			result = append(result, mapping)
		}
	}
	return result
}

// Options controls projection details.
type Options struct {
	// Header is emitted only when generating a new file.
	Header string
	// PackageName overrides the package name supplied by an adapter.
	PackageName string
}

// Result is generated Go source together with its semantic source map.
type Result struct {
	Source    []byte
	SourceMap SourceMap
}

// MetadataResult is a typed projection result with deterministic identity
// digests. Unavailable external provenance is represented explicitly.
type MetadataResult struct {
	Result
	Metadata GenerationMetadata
}
