package adapter

import (
	"fmt"
	"sort"
	"strings"
)

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
