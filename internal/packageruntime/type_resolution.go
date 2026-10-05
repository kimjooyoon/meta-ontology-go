package packageruntime

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func resolveEntityType(spec PackageSpec, name string, local map[string]string, dependencies map[string][]Export) (string, error) {
	if identity, exists := local[name]; exists {
		if identity == "" {
			return "", reject("PACKAGE_ENTITY_ID_UNKNOWN", "%s declares entity type %q without a stable ID", spec.Path, name)
		}
		return identity, nil
	}
	matches := map[string]bool{}
	for _, imported := range spec.Imports {
		for _, export := range dependencies[imported] {
			if export.Kind == "entity" && export.Name == name {
				matches[export.ID] = true
			}
		}
	}
	identities := make([]string, 0, len(matches))
	for identity := range matches {
		if identity != "" {
			identities = append(identities, identity)
		}
	}
	sort.Strings(identities)
	switch len(identities) {
	case 0:
		if len(matches) > 0 {
			return "", reject("PACKAGE_ENTITY_ID_UNKNOWN", "%s imports entity type %q without a stable ID", spec.Path, name)
		}
		return "", reject("PACKAGE_TYPE_UNKNOWN", "%s references undeclared entity type %q", spec.Path, name)
	case 1:
		return identities[0], nil
	default:
		return "", reject("PACKAGE_TYPE_AMBIGUOUS", "%s references entity type %q with multiple stable IDs: %v", spec.Path, name, identities)
	}
}

func sortExports(exports []Export) {
	sort.Slice(exports, func(i, j int) bool {
		if exports[i].Kind != exports[j].Kind {
			return exports[i].Kind < exports[j].Kind
		}
		return exports[i].Name < exports[j].Name
	})
}

func appendEntityTypeEnvironment(file *syntax.File, types map[string]string) *syntax.File {
	clone := file.Clone()
	declarations := clone.Decls
	if declarations == nil {
		declarations = clone.Declarations
	}
	present := make(map[string]bool, len(declarations))
	for _, declaration := range declarations {
		if entity, ok := declaration.(*syntax.EntityDecl); ok {
			present[entity.Name] = true
		}
	}
	names := make([]string, 0, len(types))
	for name := range types {
		if !present[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		declarations = append(declarations, &syntax.EntityDecl{Name: name, ID: types[name]})
	}
	clone.Decls = declarations
	clone.Declarations = declarations
	return clone
}
