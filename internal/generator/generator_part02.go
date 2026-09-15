package generator

// generateWithEntityFieldsSupport is intentionally package-private. Tests may
// inject the exact profile-bound SUPPORTED state, while every production
// entry point above and below remains bound to checked-in DEFERRED state.
func (g Generator) generateWithEntityFieldsSupport(input SemanticIR, previous []byte, support entityFieldsSupport) (Result, error) {
	if err := validateEntityFieldsInput(input, support); err != nil {
		return Result{}, err
	}
	if support.State == entityFieldsSupported && semanticIRHasFields(input) {
		input = prepareEntityFields(input)
	}
	ir, err := normalizeIR(input)
	if err != nil {
		return Result{}, err
	}
	markers := parsedMarkers{Slots: make(map[string]parsedSlot)}
	if len(previous) > 0 {
		markers, err = parseMarkers(previous)
		if err != nil {
			return Result{}, err
		}
		if err := validatePackage(previous, ir.Package); err != nil {
			return Result{}, err
		}
		if err := validateDeclaredRegions(ir, markers); err != nil {
			return Result{}, err
		}
	}
	blocks, order, err := g.renderBlocks(ir, markers)
	if err != nil {
		return Result{}, err
	}
	if err := validateDeclaredSlots(ir, markers, len(previous) > 0); err != nil {
		return Result{}, err
	}

	var source []byte
	if len(previous) == 0 {
		source, err = g.renderNewFile(ir, blocks, order)
	} else {
		source = patchExisting(previous, markers.Regions, blocks, order)
	}
	if err != nil {
		return Result{}, err
	}
	if err := validateGeneratedSource(source, ir.Package); err != nil {
		return Result{}, err
	}

	sourceMap, err := makeSourceMap(source, ir)
	if err != nil {
		return Result{}, err
	}
	return Result{Source: source, SourceMap: sourceMap}, nil
}
