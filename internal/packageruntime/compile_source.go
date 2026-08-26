package packageruntime

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func compileSource(spec PackageSpec, source Source) (
	SourceImage, string, []string, []EntryPlan, error,
) {
	file, diagnostics := syntax.ParseFile(source.Filename, source.Content)
	if file == nil || diagnostics.HasErrors() {
		return SourceImage{}, "", nil, nil,
			reject("PACKAGE_SOURCE_INVALID", "source %q has syntax errors", source.Filename)
	}
	if file.Package == nil || file.Namespace == nil || file.Package.Name != spec.Name {
		return SourceImage{}, "", nil, nil,
			reject("PACKAGE_HEADER_MISMATCH", "source %q does not declare package %q", source.Filename, spec.Name)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceImage{}, "", nil, nil,
			reject("PACKAGE_SOURCE_INVALID", "lower source %q: %v", source.Filename, err)
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	names, activities := sourceDeclarations(spec.Path, source.Filename, declarations)
	image := SourceImage{
		Filename: source.Filename, SourceDigest: digestValue(source.Content),
		SemanticDigest: "sha256:" + ir.StableHash(), Declarations: len(declarations),
	}
	return image, file.Namespace.Name, names, activities, nil
}
