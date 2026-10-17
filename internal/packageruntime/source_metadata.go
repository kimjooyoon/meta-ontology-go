package packageruntime

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// CompleteSourceMetadata fills omitted metadata on a package initialized from its
// manifest. Explicit assertions remain checks against the parsed source.
func CompleteSourceMetadata(pkg *PackageSpec, declaredName string, declaredImports []string) error {
	if declaredName != "" && declaredImports != nil {
		return nil
	}
	name := declaredName
	imports := map[string]bool{}
	for _, source := range pkg.Sources {
		file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(
			source.Filename, source.Content, syntax.EntityFieldsV4Support())
		if file == nil || diagnostics.HasErrors() || file.Package == nil || file.Namespace == nil {
			return fmt.Errorf("PACKAGE_SOURCE_INVALID: source %q needs valid Gooo package and namespace declarations",
				source.Filename)
		}
		if name == "" {
			name = file.Package.Name
		}
		if name != file.Package.Name {
			return fmt.Errorf("PACKAGE_HEADER_MISMATCH: source %q declares package %q; expected %q",
				source.Filename, file.Package.Name, name)
		}
		for _, item := range file.Imports {
			imports[item.Path] = true
		}
	}
	pkg.Name = name
	if declaredImports == nil {
		pkg.Imports = make([]string, 0, len(imports))
		for path := range imports {
			pkg.Imports = append(pkg.Imports, path)
		}
		sort.Strings(pkg.Imports)
	}
	return nil
}
