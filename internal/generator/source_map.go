package generator

import (
	"fmt"
	"strings"
)

func makeSourceMap(source []byte, ir SemanticIR) (SourceMap, error) {
	markers, err := parseMarkers(source)
	if err != nil {
		return SourceMap{}, err
	}
	entities, activities := sourceMapNodes(ir)
	result := SourceMap{Mappings: make([]SourceMapping, 0, len(markers.Regions))}
	for _, region := range markers.Regions {
		if err := appendSourceMapRegion(&result, source, region, entities, activities); err != nil {
			return SourceMap{}, err
		}
	}
	return result, nil
}

func sourceMapNodes(ir SemanticIR) (map[string]Entity, map[string]Activity) {
	entities := make(map[string]Entity, len(ir.Entities))
	activities := make(map[string]Activity, len(ir.Activities))
	for _, entity := range ir.Entities {
		entities[entity.ID] = entity
	}
	for _, activity := range ir.Activities {
		activities[activity.ID] = activity
	}
	return entities, activities
}

func appendSourceMapRegion(result *SourceMap, source []byte, region generatedRegion, entities map[string]Entity, activities map[string]Activity) error {
	result.Mappings = append(result.Mappings, SourceMapping{
		SemanticID: region.ID,
		Kind:       region.Kind,
		Ordinal:    len(result.Mappings),
		Source:     regionSourceSpan(region, entities, activities),
		Generated:  rangeForOffsets(source, region.Start, region.End),
	})
	if entity, ok := entities[region.ID]; ok && len(entity.Fields) > 0 {
		if err := appendEntityFieldMappings(result, source, region, entity); err != nil {
			return err
		}
	}
	return appendSlotMappings(result, source, region, activities)
}

func regionSourceSpan(region generatedRegion, entities map[string]Entity, activities map[string]Activity) SourceSpan {
	switch region.Kind {
	case "entity":
		if entity, ok := entities[region.ID]; ok {
			return entity.Source
		}
	case "activity":
		if activity, ok := activities[region.ID]; ok {
			return activity.Source
		}
	}
	return SourceSpan{}
}

func appendEntityFieldMappings(result *SourceMap, source []byte, region generatedRegion, entity Entity) error {
	ranges, err := generatedFieldRanges(source, region, entity)
	if err != nil {
		return err
	}
	profile := entityFieldsProfileMapping()
	for fieldIndex, field := range entity.Fields {
		result.Mappings = append(result.Mappings, SourceMapping{
			SemanticID:     field.ID,
			Kind:           "field",
			Ordinal:        fieldIndex,
			Source:         field.Source,
			Generated:      ranges[fieldIndex],
			ParentID:       field.Parent,
			TypeRefID:      field.TypeRefID,
			Presence:       field.Presence,
			Cardinality:    field.Cardinality,
			NameSource:     field.NameSpan,
			ProfileID:      profile.ID,
			ProfileVersion: profile.Version,
			ProfileDigest:  profile.Digest,
		})
	}
	return nil
}

func appendSlotMappings(result *SourceMap, source []byte, region generatedRegion, activities map[string]Activity) error {
	declaredSlots := make(map[string]Slot)
	if activity, ok := activities[region.ID]; ok {
		for _, declared := range activity.Slots {
			declaredSlots[declared.ID] = declared
		}
	}
	for slotIndex, slot := range region.Slots {
		declared, exists := declaredSlots[slot.ID]
		if !exists {
			return fmt.Errorf("generator: source map has stale slot identity %q", slot.ID)
		}
		result.Mappings = append(result.Mappings, SourceMapping{
			SemanticID: slot.ID,
			Kind:       "slot",
			Ordinal:    slotIndex,
			Source:     declared.Source,
			Generated:  rangeForOffsets(source, slot.Start, slot.End),
		})
	}
	return nil
}

func generatedFieldRanges(source []byte, region generatedRegion, entity Entity) ([]SourceRange, error) {
	lines := splitSourceLines(source)
	ranges := make([]SourceRange, len(entity.Fields))
	next := 0
	for fieldIndex, field := range entity.Fields {
		found, nextLine := findGeneratedFieldLine(source, lines, region, next, field)
		if !found {
			return nil, fmt.Errorf("generator: source map lost generated field %q in entity %q", field.ID, entity.ID)
		}
		ranges[fieldIndex] = nextLine.rangeValue
		next = nextLine.next
	}
	return ranges, nil
}

type generatedFieldLine struct {
	rangeValue SourceRange
	next       int
}

func findGeneratedFieldLine(source []byte, lines []sourceLine, region generatedRegion, next int, field Field) (bool, generatedFieldLine) {
	for lineIndex := next; lineIndex < len(lines); lineIndex++ {
		line := lines[lineIndex]
		if line.start < region.Start || line.end > region.End || !matchesGeneratedFieldLine(strings.TrimSpace(line.text), field) {
			continue
		}
		return true, generatedFieldLine{
			rangeValue: SourceRange{Start: positionAt(source, line.start), End: positionAt(source, line.end)},
			next:       lineIndex + 1,
		}
	}
	return false, generatedFieldLine{}
}

func matchesGeneratedFieldLine(line string, field Field) bool {
	parts := strings.Fields(line)
	return len(parts) == 2 && parts[0] == field.GoName && parts[1] == field.GoType
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
