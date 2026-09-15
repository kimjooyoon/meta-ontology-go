package generator

// ProjectionMetadataV1 is the versioned wire surface for generator output.
type ProjectionMetadataV1 struct {
	Schema     string             `json:"schema"`
	Source     []byte             `json:"source"`
	SemanticIR SemanticIR         `json:"semantic_ir"`
	SourceMap  SourceMap          `json:"source_map"`
	Metadata   GenerationMetadata `json:"metadata"`
}

// ProjectionBinding supplies independently computed identities for opt-in
// fail-closed projection verification. Empty external fields stay deferred.
type ProjectionBinding struct {
	Schema           string            `json:"schema"`
	SourceDigest     string            `json:"source_digest"`
	SemanticIRDigest string            `json:"semantic_ir_digest"`
	SourceMapDigest  string            `json:"source_map_digest"`
	EvidenceDigest   string            `json:"evidence_digest,omitempty"`
	ProvenanceDigest string            `json:"provenance_digest,omitempty"`
	Toolchain        ToolchainIdentity `json:"toolchain"`
}

// GenerationMetadata describes reproducible projection inputs and trust.
type GenerationMetadata struct {
	SourceDigest     string                `json:"source_digest"`
	SemanticIRDigest string                `json:"semantic_ir_digest"`
	SourceMapDigest  string                `json:"source_map_digest"`
	Source           BindingStatus         `json:"source"`
	SemanticIR       BindingStatus         `json:"semantic_ir"`
	Provenance       BindingStatus         `json:"provenance"`
	Evidence         EvidenceStatus        `json:"evidence"`
	Toolchain        ToolchainIdentity     `json:"toolchain"`
	Projection       ProjectionStatus      `json:"projection"`
	Authority        AuthorityLabels       `json:"authority"`
	EntityFields     *EntityFieldsMetadata `json:"entity_fields,omitempty"`
}

// EntityFieldsMetadata records the exact profile bound to a supported field
// projection. It is nil for the fieldless/deferred production path so the
// established fieldless metadata wire representation remains unchanged.
type EntityFieldsMetadata struct {
	State   string                      `json:"state"`
	Profile EntityFieldsProfileMetadata `json:"profile"`
}
type EntityFieldsProfileMetadata struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Digest  string `json:"digest"`
}

// BindingStatus identifies whether a value is available and authoritative.
type BindingStatus struct {
	Status    string `json:"status"`
	Authority string `json:"authority"`
}

// AuthorityLabels names the verifier and projection boundaries explicitly.
type AuthorityLabels struct {
	Projection string `json:"projection"`
	Verifier   string `json:"verifier"`
	Provenance string `json:"provenance"`
}

// EvidenceStatus records what this package can prove without external files.
type EvidenceStatus struct {
	Decision string   `json:"decision"`
	Refs     []string `json:"refs"`
}
