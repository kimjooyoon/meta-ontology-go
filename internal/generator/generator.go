package generator

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"sort"
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

// Generate projects ir into Go.  When previous is non-empty, only owned
// generated regions are replaced or removed.  Marker-outside text and the
// contents of stable handwritten slots are retained byte-for-byte.
func (g Generator) Generate(input SemanticIR, previous []byte) (Result, error) {
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
	}
	if err := validateRegionOwners(ir, markers); err != nil {
		return Result{}, err
	}
	if err := validateDeclaredSlots(ir, markers, len(previous) > 0); err != nil {
		return Result{}, err
	}
	blocks, order, err := g.renderBlocks(ir, markers)
	if err != nil {
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
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.gooo.go", source, parser.ParseComments); err != nil {
		return Result{}, fmt.Errorf("generator: generated source is not valid Go: %w", err)
	}

	sourceMap, err := makeSourceMap(source, ir)
	if err != nil {
		return Result{}, err
	}
	return Result{Source: source, SourceMap: sourceMap}, nil
}

func validateRegionOwners(ir SemanticIR, markers parsedMarkers) error {
	expected := make(map[string]string, len(ir.Entities)+len(ir.Activities))
	for _, entity := range ir.Entities {
		expected[entity.ID] = "entity"
	}
	for _, activity := range ir.Activities {
		expected[activity.ID] = "activity"
	}
	for _, region := range markers.Regions {
		if kind, exists := expected[region.ID]; exists && region.Kind != kind {
			return fmt.Errorf("generator: region %q changes kind from %q to %q", region.ID, region.Kind, kind)
		}
	}
	return nil
}

func validateDeclaredSlots(ir SemanticIR, markers parsedMarkers, allowRemovedRegions bool) error {
	declared := make(map[string]struct{})
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
			if _, exists := declared[slot.ID]; !exists {
				return fmt.Errorf("generator: stale slot identity %q", slot.ID)
			}
		}
	}
	for _, slot := range markers.Slots {
		owner, exists := declaredSlotOwner(ir, slot.ID)
		if !exists {
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

func makeSourceMap(source []byte, ir SemanticIR) (SourceMap, error) {
	markers, err := parseMarkers(source)
	if err != nil {
		return SourceMap{}, err
	}
	entities := make(map[string]Entity, len(ir.Entities))
	activities := make(map[string]Activity, len(ir.Activities))
	for _, entity := range ir.Entities {
		entities[entity.ID] = entity
	}
	for _, activity := range ir.Activities {
		activities[activity.ID] = activity
	}

	result := SourceMap{Mappings: make([]SourceMapping, 0, len(markers.Regions))}
	for _, region := range markers.Regions {
		var sourceSpan SourceSpan
		switch region.Kind {
		case "entity":
			if entity, ok := entities[region.ID]; ok {
				sourceSpan = entity.Source
			}
		case "activity":
			if activity, ok := activities[region.ID]; ok {
				sourceSpan = activity.Source
			}
		}
		result.Mappings = append(result.Mappings, SourceMapping{
			SemanticID: region.ID,
			Kind:       region.Kind,
			Source:     sourceSpan,
			Generated:  rangeForOffsets(source, region.Start, region.End),
		})
		for _, slot := range region.Slots {
			var slotSource SourceSpan
			if activity, ok := activities[region.ID]; ok {
				for _, declared := range activity.Slots {
					if declared.ID == slot.ID {
						slotSource = declared.Source
						break
					}
				}
			}
			result.Mappings = append(result.Mappings, SourceMapping{
				SemanticID: slot.ID,
				Kind:       "slot",
				Source:     slotSource,
				Generated:  rangeForOffsets(source, slot.Start, slot.End),
			})
		}
	}
	sort.SliceStable(result.Mappings, func(i, j int) bool {
		if result.Mappings[i].Generated.Start.Offset != result.Mappings[j].Generated.Start.Offset {
			return result.Mappings[i].Generated.Start.Offset < result.Mappings[j].Generated.Start.Offset
		}
		return result.Mappings[i].Kind < result.Mappings[j].Kind
	})
	return result, nil
}

func rangeForOffsets(source []byte, start, end int) SourceRange {
	return SourceRange{Start: positionAt(source, start), End: positionAt(source, end)}
}

func positionAt(source []byte, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	line := 1
	column := 1
	for _, value := range source[:offset] {
		if value == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return Position{Offset: offset, Line: line, Column: column}
}
