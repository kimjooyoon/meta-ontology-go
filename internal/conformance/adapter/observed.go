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

func normalizeImports(input []Import) ([]Import, error) {
	imports := append([]Import(nil), input...)
	for i := range imports {
		if strings.TrimSpace(imports[i].Path) == "" || strings.ContainsRune(imports[i].Path, 0) {
			return nil, fmt.Errorf("imports require a non-empty path")
		}
		usedBy := append([]string(nil), imports[i].UsedBy...)
		sort.Strings(usedBy)
		if hasDuplicateStrings(usedBy) {
			return nil, fmt.Errorf("import %q has duplicate used_by id", imports[i].Path)
		}
		imports[i].UsedBy = emptyStringsIfNil(usedBy)
	}
	sort.Slice(imports, func(i, j int) bool {
		if imports[i].Path != imports[j].Path {
			return imports[i].Path < imports[j].Path
		}
		return imports[i].Alias < imports[j].Alias
	})
	for i := 1; i < len(imports); i++ {
		if imports[i].Path == imports[i-1].Path && imports[i].Alias == imports[i-1].Alias {
			return nil, fmt.Errorf("duplicate import %q alias %q", imports[i].Path, imports[i].Alias)
		}
	}
	return emptyImportsIfNil(imports), nil
}

func normalizeMappings(input []Mapping) ([]Mapping, error) {
	mappings := append([]Mapping(nil), input...)
	for _, mapping := range mappings {
		if strings.TrimSpace(mapping.SemanticID) == "" || strings.TrimSpace(mapping.Kind) == "" {
			return nil, fmt.Errorf("source map entries require semantic_id and kind")
		}
		if err := validateRange(mapping.Source.Start, mapping.Source.End); err != nil {
			return nil, fmt.Errorf("source map %q source: %w", mapping.SemanticID, err)
		}
		if err := validateRange(mapping.Generated.Start, mapping.Generated.End); err != nil {
			return nil, fmt.Errorf("source map %q generated: %w", mapping.SemanticID, err)
		}
	}
	sort.Slice(mappings, func(i, j int) bool {
		if mappings[i].SemanticID != mappings[j].SemanticID {
			return mappings[i].SemanticID < mappings[j].SemanticID
		}
		return mappings[i].Kind < mappings[j].Kind
	})
	for i := 1; i < len(mappings); i++ {
		if mappings[i].SemanticID == mappings[i-1].SemanticID && mappings[i].Kind == mappings[i-1].Kind {
			return nil, fmt.Errorf("duplicate source map key %q/%q", mappings[i].SemanticID, mappings[i].Kind)
		}
	}
	return emptyMappingsIfNil(mappings), nil
}
