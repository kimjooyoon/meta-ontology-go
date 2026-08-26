package adapter

import (
	"fmt"
	"sort"
	"strings"
)

func (o Observed) normalized() (Observed, error) {
	regions, err := normalizeRegions(o.Regions)
	if err != nil {
		return Observed{}, err
	}
	slots, err := normalizeSlots(o.Slots)
	if err != nil {
		return Observed{}, err
	}
	imports, err := normalizeImports(o.Imports)
	if err != nil {
		return Observed{}, err
	}
	mappings, err := normalizeMappings(o.SourceMap)
	if err != nil {
		return Observed{}, err
	}
	delta, err := normalizeDelta(o.Delta)
	if err != nil {
		return Observed{}, err
	}
	o.Regions, o.Slots, o.Imports, o.SourceMap, o.Delta = regions, slots, imports, mappings, delta
	return o, nil
}
func normalizeRegions(input []Region) ([]Region, error) {
	regions := append([]Region(nil), input...)
	for _, region := range regions {
		if strings.TrimSpace(region.Kind) == "" || strings.TrimSpace(region.SemanticID) == "" {
			return nil, fmt.Errorf("regions require kind and semantic_id")
		}
		if err := validateRange(region.Start, region.End); err != nil {
			return nil, fmt.Errorf("region %q: %w", region.SemanticID, err)
		}
	}
	sort.Slice(regions, func(i, j int) bool {
		if regions[i].SemanticID != regions[j].SemanticID {
			return regions[i].SemanticID < regions[j].SemanticID
		}
		return regions[i].Kind < regions[j].Kind
	})
	for i := 1; i < len(regions); i++ {
		if regions[i].SemanticID == regions[i-1].SemanticID {
			return nil, fmt.Errorf("duplicate region semantic_id %q", regions[i].SemanticID)
		}
	}
	return emptyRegionsIfNil(regions), nil
}
func normalizeSlots(input []Slot) ([]Slot, error) {
	slots := append([]Slot(nil), input...)
	for _, slot := range slots {
		if strings.TrimSpace(slot.ID) == "" || strings.TrimSpace(slot.OwnerID) == "" {
			return nil, fmt.Errorf("slots require slot_id and owner_id")
		}
		if err := validateRange(slot.Start, slot.End); err != nil {
			return nil, fmt.Errorf("slot %q: %w", slot.ID, err)
		}
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i].ID < slots[j].ID })
	for i := 1; i < len(slots); i++ {
		if slots[i].ID == slots[i-1].ID {
			return nil, fmt.Errorf("duplicate slot_id %q", slots[i].ID)
		}
	}
	return emptySlotsIfNil(slots), nil
}
