package languagetest

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type specification struct {
	name, markerID, entry, expected string
}

func discover(file *syntax.File) ([]specification, map[string]Binding, error) {
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	entities := map[string]Binding{}
	for _, declaration := range declarations {
		if entity, ok := declaration.(*syntax.EntityDecl); ok {
			entities[entity.Name] = Binding{Name: entity.Name, ID: entity.ID}
		}
	}
	var specifications []specification
	seen := map[string]struct{}{}
	for _, declaration := range declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok || !strings.HasPrefix(entity.ID, MarkerPrefix) {
			continue
		}
		parts := strings.Split(strings.TrimPrefix(entity.ID, MarkerPrefix), "/")
		if len(parts) != 3 || parts[0] == "" || parts[1] != "output" || parts[2] == "" {
			return nil, nil, fmt.Errorf("test marker %q must use %s<entry>/output/<entity>", entity.ID, MarkerPrefix)
		}
		if _, duplicate := seen[entity.Name]; duplicate {
			return nil, nil, fmt.Errorf("test name %q is duplicated", entity.Name)
		}
		expected, exists := entities[parts[2]]
		if !exists || strings.HasPrefix(expected.ID, MarkerPrefix) {
			return nil, nil, fmt.Errorf("expected output entity %q is not a runtime entity", parts[2])
		}
		seen[entity.Name] = struct{}{}
		specifications = append(specifications, specification{
			name: entity.Name, markerID: entity.ID, entry: parts[0], expected: parts[2],
		})
	}
	return specifications, entities, nil
}
