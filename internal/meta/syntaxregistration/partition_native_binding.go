package syntaxregistration

import (
	"fmt"
	"go/ast"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

type partitionNativeBinding struct {
	root                     *ast.CompositeLit
	families, packages, meta []*ast.CompositeLit
}

func bindPartitionNative(source *goSource, registry languagesyntax.Registry) (partitionNativeBinding, error) {
	root, err := source.function("expectedRegistry")
	if err != nil {
		return partitionNativeBinding{}, err
	}
	literal, err := registryReturnLiteral(root)
	if err != nil {
		return partitionNativeBinding{}, err
	}
	rows, err := readProfileNativeCases(source)
	if err != nil {
		return partitionNativeBinding{}, err
	}
	definitions := make([]languagesyntax.CaseDefinition, len(rows))
	for i, row := range rows {
		definitions[i] = row.definition
	}
	families, familyOrigins, err := readNativeCaseSources(source)
	if err != nil {
		return partitionNativeBinding{}, err
	}
	packages, packageOrigins, err := readNativePackageUnits(source)
	if err != nil {
		return partitionNativeBinding{}, err
	}
	meta, err := source.registryFieldLeaves(root, "MetaSources", "string", false)
	if err != nil {
		return partitionNativeBinding{}, err
	}
	var metaPaths []string
	for _, list := range meta {
		paths, err := nativeMetaSourceValues(source, list)
		if err != nil {
			return partitionNativeBinding{}, err
		}
		metaPaths = append(metaPaths, paths...)
	}
	if !reflect.DeepEqual(definitions, registry.Cases) || !reflect.DeepEqual(families, registry.CaseSources) ||
		!reflect.DeepEqual(packages, registry.PackageUnits) || !reflect.DeepEqual(metaPaths, registry.MetaSources) {
		return partitionNativeBinding{}, fmt.Errorf("native partition registry differs from the corpus")
	}
	if err := bindPartitionRegistryRemainder(source, root, literal, registry); err != nil {
		return partitionNativeBinding{}, err
	}
	return partitionNativeBinding{literal, familyOrigins, packageOrigins, meta}, nil
}

func bindPartitionRegistryRemainder(source *goSource, root *ast.FuncDecl, literal *ast.CompositeLit, registry languagesyntax.Registry) error {
	seen := map[string]bool{}
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return fmt.Errorf("native registry requires named fields")
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok || seen[key.Name] || !reflect.ValueOf(registry).FieldByName(key.Name).IsValid() {
			return fmt.Errorf("native registry field is unknown or repeated")
		}
		seen[key.Name] = true
	}
	schema, err := registryLiteralField(literal, "Schema")
	if err != nil {
		return err
	}
	value, err := profileConstantValue(source, schema, map[string]bool{})
	if err != nil {
		return err
	}
	var name string
	if err := assignProfileCaseValue(reflect.ValueOf(&name).Elem(), value); err != nil {
		return err
	}
	if name != registry.Schema {
		return fmt.Errorf("native registry schema differs from the corpus")
	}
	origins, err := source.registryFieldOrigins(root, "ProjectionUnits", "ProjectionDefinition", true)
	if err != nil {
		return err
	}
	var units []languagesyntax.ProjectionDefinition
	for _, leaf := range origins.leaves {
		for _, element := range leaf.Elts {
			row, err := origins.literal(element, "ProjectionDefinition")
			if err != nil {
				return err
			}
			var unit languagesyntax.ProjectionDefinition
			if err := readPartitionRecord(source, row, &unit); err != nil {
				return err
			}
			units = append(units, unit)
		}
	}
	if !reflect.DeepEqual(units, registry.ProjectionUnits) {
		return fmt.Errorf("native projection consumers differ from the corpus")
	}
	return nil
}
