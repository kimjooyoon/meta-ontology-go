package packageruntime

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type compiledPackage struct {
	image      PackageImage
	activities []EntryPlan
}

type parsedPackageSource struct {
	source Source
	file   *syntax.File
	names  []string
}

func compilePackage(spec PackageSpec, dependencies map[string][]Export) (compiledPackage, error) {
	compiled := compiledPackage{image: PackageImage{
		Path: spec.Path, Name: spec.Name, Imports: append([]string(nil), spec.Imports...),
	}}
	declarations := map[string]bool{}
	var exports []Export
	var sources []parsedPackageSource
	sourceImports := map[string]bool{}
	for _, source := range spec.Sources {
		file, names, sourceExports, activities, err := compileSource(spec, source)
		if err != nil {
			return compiledPackage{}, err
		}
		if compiled.image.Namespace != "" && compiled.image.Namespace != file.Namespace.Name {
			return compiledPackage{}, reject("PACKAGE_NAMESPACE_MISMATCH", "package %q", spec.Path)
		}
		compiled.image.Namespace = file.Namespace.Name
		for _, name := range names {
			if declarations[name] {
				return compiledPackage{}, reject("PACKAGE_DECLARATION_DUPLICATE", "%s:%s", spec.Path, name)
			}
			declarations[name] = true
		}
		exports = append(exports, sourceExports...)
		sources = append(sources, parsedPackageSource{source: source, file: file, names: names})
		compiled.image.Declarations += len(names)
		compiled.activities = append(compiled.activities, activities...)
		for _, importDecl := range file.Imports {
			sourceImports[importDecl.Path] = true
		}
	}
	if !sameImportSet(spec.Imports, sourceImports) {
		return compiledPackage{}, reject("PACKAGE_SOURCE_IMPORT_MISMATCH", "%s source imports do not match the workspace manifest", spec.Path)
	}
	entityTypes := make(map[string]string)
	for _, export := range exports {
		if export.Kind == "entity" {
			entityTypes[export.Name] = export.ID
		}
	}
	typeEnvironment := make(map[string]string)
	for index := range exports {
		if exports[index].Kind != "activity" {
			continue
		}
		resolvedInputs := make([]string, len(exports[index].InputTypes))
		for typeIndex, name := range exports[index].InputTypes {
			resolved, err := resolveEntityType(spec, name, entityTypes, dependencies)
			if err != nil {
				return compiledPackage{}, err
			}
			resolvedInputs[typeIndex] = resolved
			typeEnvironment[name] = resolved
		}
		resolvedOutput, err := resolveEntityType(spec, exports[index].OutputType, entityTypes, dependencies)
		if err != nil {
			return compiledPackage{}, err
		}
		typeEnvironment[exports[index].OutputType] = resolvedOutput
		exports[index].InputTypes = resolvedInputs
		exports[index].OutputType = resolvedOutput
	}
	for _, source := range sources {
		bindings, err := resolveImportedBindings(spec, source.file, exports, dependencies)
		if err != nil {
			return compiledPackage{}, err
		}
		compiled.image.Bindings = append(compiled.image.Bindings, bindings...)
	}
	for _, source := range sources {
		fileForLowering := source.file.Clone()
		fileForLowering.Bindings = localBindings(fileForLowering.Bindings)
		fileWithTypes := appendEntityTypeEnvironment(fileForLowering, typeEnvironment)
		ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), fileWithTypes, bidir.EntityFieldsV4Support())
		if err != nil {
			return compiledPackage{}, reject("PACKAGE_SOURCE_INVALID", "lower source %q: %v", source.source.Filename, err)
		}
		compiled.image.Sources = append(compiled.image.Sources, SourceImage{
			Filename: source.source.Filename, SourceDigest: digestValue(source.source.Content),
			SemanticDigest: "sha256:" + ir.StableHash(), Declarations: len(source.names),
		})
	}
	sortExports(exports)
	compiled.image.Exports = exports
	compiled.image.SemanticDigest = digestValue(struct {
		Path, Namespace string
		Sources         []SourceImage
		Exports         []Export
		Bindings        []PackageBinding
	}{spec.Path, compiled.image.Namespace, compiled.image.Sources, compiled.image.Exports, compiled.image.Bindings})
	return compiled, nil
}

func sameImportSet(manifestImports []string, sourceImports map[string]bool) bool {
	if len(sourceImports) == 0 {
		return true
	}
	if len(manifestImports) != len(sourceImports) {
		return false
	}
	for _, importPath := range manifestImports {
		if !sourceImports[importPath] {
			return false
		}
	}
	return true
}
