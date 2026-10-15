package packageruntime

import (
	"path"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func sourceImportAliases(spec PackageSpec, file *syntax.File) (map[string]string, error) {
	imports := make(map[string]string, len(file.Imports))
	for _, declaration := range file.Imports {
		alias := declaration.Alias
		if alias == "" {
			alias = path.Base(declaration.Path)
		}
		if previous, exists := imports[alias]; exists && previous != declaration.Path {
			return nil, reject("PACKAGE_IMPORT_ALIAS_AMBIGUOUS", "%s uses alias %q for both %q and %q",
				spec.Path, alias, previous, declaration.Path)
		}
		imports[alias] = declaration.Path
	}
	return imports, nil
}

func resolveSourceEntityType(spec PackageSpec, file *syntax.File, spelling string,
	local map[string]string, dependencies map[string][]Export) (string, error) {
	alias, member, qualified := strings.Cut(spelling, ".")
	if !qualified {
		return resolveEntityType(spec, spelling, local, dependencies)
	}
	imports, err := sourceImportAliases(spec, file)
	if err != nil {
		return "", err
	}
	imported, ok := imports[alias]
	if !ok {
		return "", reject("PACKAGE_TYPE_IMPORT_UNKNOWN", "%s uses undeclared import alias %q in type %q",
			spec.Path, alias, spelling)
	}
	selected := spec
	selected.Imports = []string{imported}
	identity, err := resolveEntityType(selected, member, nil, dependencies)
	if err != nil {
		return "", reject("PACKAGE_IMPORTED_TYPE_INVALID", "%s type %q from %q: %v", spec.Path, spelling, imported, err)
	}
	return identity, nil
}
