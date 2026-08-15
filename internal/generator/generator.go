package generator

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
)

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

func validateDeclaredSlots(ir SemanticIR, markers parsedMarkers, allowRemovedRegions bool) error {
	declared := make(map[string]struct{})
	active := make(map[string]struct{})
	for _, activity := range ir.Activities {
		for _, slot := range activity.Slots {
			declared[slot.ID] = struct{}{}
		}
	}
	for _, region := range markers.Regions {
		if allowRemovedRegions && !hasActivity(ir, region.ID) {
			continue
		}
		for _, slot := range region.Slots {
			active[slot.ID] = struct{}{}
		}
	}
	for id := range active {
		if _, exists := declared[id]; !exists {
			return fmt.Errorf("generator: stale slot identity %q", id)
		}
	}
	for _, slot := range markers.Slots {
		owner, declared := declaredSlotOwner(ir, slot.ID)
		if !declared {
			continue
		}
		if owner != slot.RegionID {
			return fmt.Errorf("generator: slot %q changes region owner from %q to %q", slot.ID, slot.RegionID, owner)
		}
		if slot.RegionKind != "activity" {
			return fmt.Errorf("generator: slot %q belongs to non-activity region kind %q", slot.ID, slot.RegionKind)
		}
	}
	return nil
}

func declaredSlotOwner(ir SemanticIR, slotID string) (string, bool) {
	for _, activity := range ir.Activities {
		for _, slot := range activity.Slots {
			if slot.ID == slotID {
				return activity.ID, true
			}
		}
	}
	return "", false
}

func hasActivity(ir SemanticIR, id string) bool {
	for _, activity := range ir.Activities {
		if activity.ID == id {
			return true
		}
	}
	return false
}

func validatePackage(source []byte, expected string) error {
	file, err := parser.ParseFile(token.NewFileSet(), "previous.go", source, parser.PackageClauseOnly)
	if err != nil {
		return fmt.Errorf("generator: previous source has no readable package clause: %w", err)
	}
	if file.Name.Name != expected {
		return fmt.Errorf("generator: previous package %q does not match semantic package %q", file.Name.Name, expected)
	}
	return nil
}

func patchExisting(previous []byte, regions []generatedRegion, blocks map[string][]byte, order []string) []byte {
	var output bytes.Buffer
	cursor := 0
	present := make(map[string]struct{}, len(regions))
	for _, region := range regions {
		output.Write(previous[cursor:region.Start])
		if block, exists := blocks[region.ID]; exists {
			output.Write(block)
			present[region.ID] = struct{}{}
		}
		cursor = region.End
	}
	output.Write(previous[cursor:])

	for _, id := range order {
		if _, exists := present[id]; exists {
			continue
		}
		appendGeneratedBlock(&output, blocks[id])
	}
	return output.Bytes()
}

func appendGeneratedBlock(output *bytes.Buffer, block []byte) {
	value := output.Bytes()
	if len(value) > 0 && value[len(value)-1] != '\n' {
		output.WriteByte('\n')
	}
	if output.Len() > 0 {
		output.WriteByte('\n')
	}
	output.Write(block)
}
