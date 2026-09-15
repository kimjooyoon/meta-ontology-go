package generator

const projectionMetadataSchemaV1 = "gooo-generator/v1"
const projectionBindingSchemaV1 = "gooo-generator-binding/v1"

// GenerateWithMetadata is a companion API; Generate and GenerateFrom remain
// unchanged for existing callers.
func GenerateWithMetadata(ir SemanticIR, previous []byte) (MetadataResult, error) {
	result, err := Generate(ir, previous)
	if err != nil {
		return MetadataResult{}, err
	}
	return metadataResult(result, ir), nil
}

// GenerateProjectionV1 returns a versioned, canonicalizable result surface.
func GenerateProjectionV1(ir SemanticIR, previous []byte) (ProjectionMetadataV1, error) {
	return generateProjectionV1(New(Options{}), ir, previous)
}
func generateProjectionV1(generator Generator, ir SemanticIR, previous []byte) (ProjectionMetadataV1, error) {
	return generateProjectionV1WithEntityFieldsSupport(generator, ir, previous, checkedEntityFieldsSupport())
}

// generateProjectionV1WithEntityFieldsSupport is package-private so focused
// tests can prove the exact profile-bound SUPPORTED branch without exposing a
// caller-selectable production activation surface.
func generateProjectionV1WithEntityFieldsSupport(generator Generator, ir SemanticIR, previous []byte, support entityFieldsSupport) (ProjectionMetadataV1, error) {
	if err := validateEntityFieldsInput(ir, support); err != nil {
		return ProjectionMetadataV1{}, err
	}
	working := ir
	if support.State == entityFieldsSupported && semanticIRHasFields(ir) {
		working = prepareEntityFields(ir)
	}
	normalized, err := normalizeIR(working)
	if err != nil {
		return ProjectionMetadataV1{}, err
	}
	result, err := generator.generateWithEntityFieldsSupport(normalized, previous, support)
	if err != nil {
		return ProjectionMetadataV1{}, err
	}
	metadata := metadataResultWithEntityFieldsSupport(result, normalized, support)
	return ProjectionMetadataV1{
		Schema:     projectionMetadataSchemaV1,
		Source:     append([]byte(nil), result.Source...),
		SemanticIR: normalized,
		SourceMap:  result.SourceMap,
		Metadata:   metadata.Metadata,
	}, nil
}
