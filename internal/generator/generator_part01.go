package generator

// Project is the convenience entry point for a projection with optional
// previous source.  It is the strongly typed API used by compiler adapters.
func Project(ir SemanticIR, previous []byte) (Result, error) {
	return New(Options{}).Generate(ir, previous)
}

// Generate projects a local SemanticIR with optional previous source.
func Generate(ir SemanticIR, previous []byte) (Result, error) {
	return New(Options{}).Generate(ir, previous)
}

// GenerateFrom is a compatibility-friendly one-shot API. Inputs may be a
// SemanticIR, a SemanticIRProvider, or a structurally compatible semantic
// graph supplied by an adapter. The return source and source map are kept
// separate for callers that write the source directly to disk.
func GenerateFrom(input any, options Options) ([]byte, SourceMap, error) {
	ir, err := adaptInput(input)
	if err != nil {
		return nil, SourceMap{}, err
	}
	if options.PackageName != "" {
		ir.Package = options.PackageName
	}
	result, err := New(options).Generate(ir, nil)
	if err != nil {
		return nil, SourceMap{}, err
	}
	return result.Source, result.SourceMap, nil
}

// GenerateFromProjectionV1 adapts a typed or reflective input through the
// same strict compatibility path as GenerateFrom and returns versioned,
// read-only projection metadata. External evidence remains deferred.
func GenerateFromProjectionV1(input any, options Options) (ProjectionMetadataV1, error) {
	ir, err := adaptInput(input)
	if err != nil {
		return ProjectionMetadataV1{}, err
	}
	if options.PackageName != "" {
		ir.Package = options.PackageName
	}
	return generateProjectionV1(New(options), ir, nil)
}

// Generate projects ir into Go.  When previous is non-empty, only owned
// generated regions are replaced or removed.  Marker-outside text and the
// contents of stable handwritten slots are retained byte-for-byte.
func (g Generator) Generate(input SemanticIR, previous []byte) (Result, error) {
	return g.generateWithEntityFieldsSupport(input, previous, checkedEntityFieldsSupport())
}
