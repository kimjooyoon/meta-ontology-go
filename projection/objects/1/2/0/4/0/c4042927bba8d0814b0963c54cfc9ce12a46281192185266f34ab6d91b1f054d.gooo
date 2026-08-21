package generator

import (
	"fmt"
)

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
