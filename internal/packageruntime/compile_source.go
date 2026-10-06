package packageruntime

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func compileSource(spec PackageSpec, source Source) (*syntax.File, []string, []Export, []EntryPlan, error) {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(source.Filename, source.Content, syntax.EntityFieldsV3Support())
	if file == nil || diagnostics.HasErrors() {
		return nil, nil, nil, nil,
			reject("PACKAGE_SOURCE_INVALID", "source %q has syntax errors", source.Filename)
	}
	if file.Package == nil || file.Namespace == nil || file.Package.Name != spec.Name {
		return nil, nil, nil, nil,
			reject("PACKAGE_HEADER_MISMATCH", "source %q does not declare package %q", source.Filename, spec.Name)
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	names, exports, activities := sourceDeclarations(spec.Path, source.Filename, declarations)
	return file, names, exports, activities, nil
}
